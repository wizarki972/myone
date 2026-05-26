package services

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/fldir"
	"github.com/wizarki972/myone/internal/utils/logger"
	"github.com/wizarki972/myone/internal/utils/pkg"
)

const HYPRCTL_MONITORS_CMD = "hyprctl -j monitors"

var (
	bctlRegExp        = regexp.MustCompile(`Device '([^']+)' of class 'backlight'`)
	i2cbusMatch       = regexp.MustCompile(`/dev/i2c-(\d+)`)
	drmConnectorMatch = regexp.MustCompile(`DRM connector:\s*((card\d+)-([\w-]+))`)
	serialNumberMatch = regexp.MustCompile(`Serial number:\s+([\w-]+)`)
)

type DisplayType int

const (
	Backlight DisplayType = iota
	DDC
	AppleDisplay
	Unknown
	Invalid
)

type hyprMonitor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Make        string `json:"make"`
}

type Monitor struct {
	// static
	Name        string
	cardName    string
	BusNum      string
	SerialNum   string
	Backlight   string
	DisplayType DisplayType

	// non-static
	maxBrightness     float64
	currentBrightness float64
	mu                sync.RWMutex
}

func (monitor *Monitor) setBrightness(value float64, userConfig *config.Config, loggBook *logger.LogBook) {
	var command string
	handledefault := func() {
		loggBook.EnterLogAndPrint("Performing default action", logger.LogTypes.Warning, nil)
		command = fmt.Sprintf("brightnessctl s %.0f", value)

	}
	switch monitor.DisplayType {
	case Backlight:
		if len(monitor.Backlight) > 0 {
			command = fmt.Sprintf("brightnessctl --device %s s %.0f", monitor.Backlight, value)
			break
		}
		loggBook.EnterLogAndPrint(fmt.Sprintf("For display %s no backlight was found, but the display type is backlight.", monitor.Name), logger.LogTypes.Warning, nil)
		handledefault()
	case DDC:
		if len(monitor.BusNum) > 0 {
			command = fmt.Sprintf("ddcutil -b %s setvcp 10 %.0f", monitor.BusNum, value)
			break
		}
		loggBook.EnterLogAndPrint(fmt.Sprintf("For display %s no bus number was found, but the display type is DDC ", monitor.Name), logger.LogTypes.Warning, nil)
		handledefault()
	case AppleDisplay:
		command = fmt.Sprintf("asdbctl set %.0f", value)

		if !userConfig.Experimental.UseSerialIDForASD {
			break
		}
		if len(monitor.SerialNum) > 0 {
			command = fmt.Sprintf("asdbctl -s %s set %.0f", monitor.SerialNum, value)
			break
		}
		loggBook.EnterLogAndPrint(fmt.Sprintf("For display %s no serial number was found, but the display type is Apple Studio Display ", monitor.Name), logger.LogTypes.Warning, nil)
		// handledefault()
	default:
		handledefault()
	}

	if len(command) == 0 {
		loggBook.EnterLogAndPrint("Something thats not supposed to happen happened, since no command was chosen, the service cannot change the brightness.", logger.LogTypes.Warning, nil)
	}

	monitor.mu.Lock()
	if _, err := cmds.ExecCommand(command, false, false); err != nil {
		monitor.mu.Unlock()
		loggBook.EnterLogAndPrint("Error in executing this command --> "+command, logger.LogTypes.Warning, nil)
		return
	}
	monitor.currentBrightness = value
	monitor.mu.Unlock()
	loggBook.EnterLogAndPrint(fmt.Sprintf("%s --> brightness changed from %.2f to %.2f", monitor.Name, monitor.currentBrightness, value), logger.LogTypes.Info, nil)
}

type MonitorManager struct {
	loggBook   *logger.LogBook
	userConfig *config.Config
	mu         sync.RWMutex

	// socket paths...
	hyprlandSocket2 string
	mmSocket        string

	// dependency checks...
	ddcutilPresent bool
	asdbctlPresent bool

	// data...
	monitors                 map[string]*Monitor
	brightnessChangeRequests chan map[string]float64
	quit                     chan string
}

func NewMonitorManager(loggBook *logger.LogBook, userConfig *config.Config) *MonitorManager {
	if userConfig.Experimental.UseSerialIDForASD {
		loggBook.EnterLogAndPrint("Experimental :: Using serial ID of Monitors. Helps with multiple Apple Studio Displays", logger.LogTypes.Info, nil)
	}

	runtimeDir, err := fldir.GetRuntimeDir()
	if err != nil {
		loggBook.EnterLogAndPrint("Failed to get a proper runtime directory.", logger.LogTypes.Error, err)
	}
	hyprlandInstanceSign := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if len(strings.TrimSpace(runtimeDir)) == 0 {
		loggBook.EnterLogAndPrint("Cannot get XDG Runtime Directory environment variable.", logger.LogTypes.Error, errors.New("cannot get XDG Runtime Directory environment variable"))
	}
	if len(strings.TrimSpace(hyprlandInstanceSign)) == 0 {
		loggBook.EnterLogAndPrint("Cannot get Hyprland Instance Signature environment variable.", logger.LogTypes.Error, errors.New("cannot get Hyprland Instance Signature environment variable"))
	}

	return &MonitorManager{
		userConfig: userConfig,
		loggBook:   loggBook,

		mmSocket:        filepath.Join(runtimeDir, "myone", "monitor.sock"),
		hyprlandSocket2: filepath.Join(runtimeDir, "hypr", hyprlandInstanceSign, ".socket2.sock"),

		ddcutilPresent: pkg.IsPkgInstalled("ddcutil"),
		asdbctlPresent: pkg.IsPkgInstalled("asdbctl"),

		monitors:                 make(map[string]*Monitor),
		brightnessChangeRequests: make(chan map[string]float64, 10),
		quit:                     make(chan string),
	}
}

// BELOW CODE FOR - BRIGHTNESS REQUESTS

// handles new  brightness change requets
func (mm *MonitorManager) brightnessRequestHandler() {
	for {
		select {
		case req := <-mm.brightnessChangeRequests:
			mm.mu.RLock()
			for monitorName, value := range req {
				monitor, ok := mm.monitors[monitorName]
				// if monitor not found...
				if !ok {
					mm.loggBook.EnterLogAndPrint("Monitor not recognized (or) not found.", logger.LogTypes.Warning, nil)
					break
				}

				// what if the information is lacking/display invalid
				if monitor.DisplayType == Invalid {
					mm.loggBook.EnterLogAndPrint("Monitor "+monitorName+" is invalid.", logger.LogTypes.Warning, nil)
					break
				}

				// what if the brightness value is less thean/equal to zero...
				if value > 0 {
					monitor.setBrightness(value, mm.userConfig, mm.loggBook)
				} else {
					mm.loggBook.EnterLogAndPrint("[ERROR] Invalid brightness value.", logger.LogTypes.Warning, nil)
					break
				}
			}
			mm.mu.RUnlock()
		case <-mm.quit:
			return
		}
	}
}

// adds new brightness change requests
func (mm *MonitorManager) addBrightnessRequest(monitorName, value string) {
	if len(strings.TrimSpace(monitorName)) == 0 || len(strings.TrimSpace(value)) == 0 {
		mm.loggBook.EnterLogAndPrint("Invalid brightness change values received.", logger.LogTypes.Warning, nil)
		return
	}

	mm.mu.RLock()
	monitor, ok := mm.monitors[monitorName]
	mm.mu.RUnlock()

	if !ok {
		mm.loggBook.EnterLogAndPrint("Invalid monitor name received - "+monitorName, logger.LogTypes.Warning, nil)
		return
	}

	monitor.mu.RLock()
	currentBrightness := monitor.currentBrightness
	maxBrightness := monitor.maxBrightness
	monitor.mu.RUnlock()

	if monitor.DisplayType == Invalid {
		mm.loggBook.EnterLogAndPrint("Invalid display type.", logger.LogTypes.Warning, nil)
		return
	}

	trimmed := strings.TrimPrefix(strings.TrimSpace(value), "%")
	if len(trimmed) == 0 {
		mm.loggBook.EnterLogAndPrint("Invalid brightness change request received. monitor name:"+monitorName+", value:"+value, logger.LogTypes.Warning, nil)
		return
	}

	getFloatValue := func(startPosition int) (float64, error) {
		if startPosition >= len(trimmed) || startPosition < 0 {
			mm.loggBook.EnterLogAndPrint("Start position out of bounds.", logger.LogTypes.Warning, nil)
			return -1, errors.New("start position out of bounds")
		}
		floatValue, err := strconv.ParseFloat(trimmed[startPosition:], 64)
		if err != nil {
			mm.loggBook.EnterLogAndPrint("Cannot convert value to float64.", logger.LogTypes.Warning, nil)
			return -1, err
		}
		if floatValue <= 0 {
			mm.loggBook.EnterLogAndPrint("Value less than or equal to 0 is received.", logger.LogTypes.Warning, nil)
			return -1, errors.New("Value less than or equal to 0 is received.")
		}
		return (floatValue / 100) * maxBrightness, nil
	}

	firstRune := []rune(trimmed)[0]
	var floatValue float64
	var err error
	switch {
	case firstRune == '+':
		floatValue, err = getFloatValue(1)
		if err != nil {
			return
		}
		floatValue = currentBrightness + floatValue
	case firstRune == '-':
		floatValue, err = getFloatValue(1)
		if err != nil {
			return
		}
		floatValue = currentBrightness - floatValue
	case unicode.IsDigit(firstRune):
		floatValue, err = getFloatValue(0)
		if err != nil {
			return
		}
	default:
		mm.loggBook.EnterLogAndPrint("Invalid brightness change request received. monitor name:"+monitorName+", value:"+value, logger.LogTypes.Warning, nil)
		return
	}
	mm.brightnessChangeRequests <- map[string]float64{monitorName: floatValue}
}

// BELOW CODE FOR - SOCKET LISTENERS & SERVICE STARTER

func (mm *MonitorManager) HyprlandIPCListener() {
	mm.loggBook.EnterLogAndPrint("Starting Hyprland IPC listener...", logger.LogTypes.Info, nil)
	conn, err := net.Dial("unix", mm.hyprlandSocket2)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Cannot dial hyprland socket2 - "+mm.hyprlandSocket2, logger.LogTypes.Error, err)
		return
	}
	defer conn.Close()

	mm.loggBook.EnterLogAndPrint("Monitoring Hyprland IPC socket -> "+mm.hyprlandSocket2, logger.LogTypes.Info, nil)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "monitorremoved") || strings.HasPrefix(line, "monitoradded") {
			mm.loggBook.EnterLogAndPrint("Change detected in the number of monitors. Updating monitors information.", logger.LogTypes.Info, nil)
			mm.Discover()
		}
	}
}

func (mm *MonitorManager) StartService() {
	if len(mm.monitors) == 0 {
		mm.Discover()
	}
	go mm.HyprlandIPCListener()
	go mm.brightnessRequestHandler()

	mm.requestListener()
}

// listens for all kinds of monitor releated requests from the socket, but currently only for brightness change requests.
func (mm *MonitorManager) requestListener() {
	os.Remove(mm.mmSocket)

	listener, err := net.Listen("unix", mm.mmSocket)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to listen from socket - "+mm.mmSocket, logger.LogTypes.Error, err)
	}
	defer listener.Close()

	mm.loggBook.EnterLogAndPrint("Listening for requets...", logger.LogTypes.Info, nil)
	for {
		c, err := listener.Accept()
		if err != nil {
			mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		}

		go func(conn net.Conn) {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				input := scanner.Text()
				mm.loggBook.EnterLogAndPrint("Received ==> "+input, logger.LogTypes.Info, nil)
				args := strings.Split(strings.TrimSpace(input), ">>")
				if len(args) == 0 {
					mm.loggBook.EnterLogAndPrint("Invalid request. "+input, logger.LogTypes.Warning, nil)
					return
				}

				switch args[0] {
				case "brightness":
					if len(args) != 3 {
						mm.loggBook.EnterLogAndPrint("Invalid request. "+input, logger.LogTypes.Warning, nil)
						return
					}
					mm.addBrightnessRequest(args[1], args[2])
				default:
					mm.loggBook.EnterLogAndPrint("Invalid request. "+input, logger.LogTypes.Warning, nil)
					return
				}
			}
		}(c)
	}
}

// BELOW CODE IS FOR - DISCOVER MONITORS AND GET BRIGHTNESS VALUES

// Discover all available monitors
func (mm *MonitorManager) Discover() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.loggBook.EnterLogAndPrint("Scanning for all monitors...", logger.LogTypes.Info, nil)
	mm.monitors = make(map[string]*Monitor)
	mm.prepareMonitorsSlice()
	mm.fillMonitorValues()
	mm.getBrightnessValues()
}

// prepare monitor slice with all compositor recognized monitors...
func (mm *MonitorManager) prepareMonitorsSlice() {
	// getting compositor recognized monitors...
	output, err := cmds.ExecCommandBytes(HYPRCTL_MONITORS_CMD, true)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Error while executing command - "+HYPRCTL_MONITORS_CMD, logger.LogTypes.Error, err)
	}
	tempMonitors := make([]hyprMonitor, 0)
	if err := json.Unmarshal(output, &tempMonitors); err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to parse json values from hyprctl "+HYPRCTL_MONITORS_CMD, logger.LogTypes.Error, err)
	}

	// preparing monitors slice...
	for _, monitor := range tempMonitors {
		mm.monitors[monitor.Name] = &Monitor{
			Name:              monitor.Name,
			DisplayType:       Unknown,
			maxBrightness:     -1,
			currentBrightness: -1,
		}

		mm.monitors[monitor.Name].mu.Lock()
		// Apple Check...
		preprocessedDescription := strings.ReplaceAll(strings.ToLower(monitor.Description), " ", "")
		if strings.Contains(preprocessedDescription, "apple") || strings.Contains(preprocessedDescription, "studiodisplay") {
			mm.monitors[monitor.Name].DisplayType = AppleDisplay
		} else {
			preprocessedMake := strings.ReplaceAll(strings.ToLower(monitor.Make), " ", "")
			if strings.Contains(preprocessedMake, "apple") || strings.Contains(preprocessedMake, "applecomputerinc") {
				mm.monitors[monitor.Name].DisplayType = AppleDisplay
			}
		}
		mm.monitors[monitor.Name].mu.Unlock()
	}
}

// fill monitors slice with necessary values...
func (mm *MonitorManager) fillMonitorValues() {
	// brightnessctl i output...
	bctlOut, err := cmds.ExecCommand("brightnessctl i", false, true)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Error cannot run command - 'brightnessctl i'", logger.LogTypes.Error, err)
	}
	var bctlDevices []string
	for block := range strings.SplitSeq(bctlOut, "\n\n") {
		for line := range strings.SplitSeq(block, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "Device") {
				matches := bctlRegExp.FindStringSubmatch(line)
				if len(matches) > 1 {
					bctlDevices = append(bctlDevices, matches[1])
				}
				continue
			}
		}
	}

	// ddcutil detect output
	ddcUtilOut, err := cmds.ExecCommand("ddcutil detect", false, true)
	if err != nil {
		if !mm.ddcutilPresent {
			mm.loggBook.EnterLogAndPrint("ddcutil dependency not found", logger.LogTypes.Error, err)
			return
		}
		mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

ddcUtilBlockLoop:
	for block := range strings.SplitSeq(ddcUtilOut, "\n\n") {
		// skipping failure messages & invalid blocks...
		var failedFound = false
		var filtered []string
		for line := range strings.SplitSeq(block, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "Failed") {
				failedFound = true
				continue
			}
			if strings.HasPrefix(trimmed, "Invalid display") {
				continue ddcUtilBlockLoop
			}
			if !failedFound {
				break
			}
			filtered = append(filtered, line)
		}
		if failedFound {
			block = strings.Join(filtered, "\n")
		}

		// getting card name and monitor...
		matches := drmConnectorMatch.FindStringSubmatch(block)
		if len(matches) != 4 {
			continue
		}
		monitor, ok := mm.monitors[matches[3]]
		if !ok {
			mm.loggBook.EnterLogAndPrint("Skipped "+matches[1]+", this monitor is not recognized by the compositor.", logger.LogTypes.Warning, err)
			continue
		}
		monitor.mu.Lock()
		monitor.cardName = matches[1]

		drmPath := filepath.Join("/sys/class/drm", monitor.cardName)
		if fldir.IsPathExist(drmPath) {
			// apple display - serial number
			if monitor.DisplayType == AppleDisplay {
				if mm.userConfig.Experimental.UseSerialIDForASD {
					serialMatches := serialNumberMatch.FindStringSubmatch(block)
					if len(serialMatches) == 2 {
						monitor.SerialNum = serialMatches[1]
					} else {
						monitor.DisplayType = Invalid
						mm.loggBook.EnterLogAndPrint("Failed to get proper serial number for "+monitor.cardName, logger.LogTypes.Warning, nil)
					}
				}
				monitor.mu.Unlock()
				continue
			}

			// eDP backlight matching...
			if strings.Contains(monitor.Name, "eDP") && len(bctlDevices) > 0 {
				for _, backlightName := range bctlDevices {
					backlightPath := filepath.Join(drmPath, backlightName)
					if fldir.IsPathExist(backlightPath) {
						monitor.DisplayType = Backlight
						monitor.Backlight = backlightName
						monitor.mu.Unlock()
						continue ddcUtilBlockLoop
					}
				}
			}

			// ddc - bus number...
			monitor.DisplayType = DDC
			i2cbusMatches := i2cbusMatch.FindStringSubmatch(block)
			if len(i2cbusMatches) == 2 {
				monitor.BusNum = i2cbusMatches[1]
			} else {
				monitor.DisplayType = Invalid
			}
			monitor.mu.Unlock()
			continue
		}
		monitor.DisplayType = Invalid
		monitor.mu.Unlock()
	}
}

// get the monitors current and max brightness...
func (mm *MonitorManager) getBrightnessValues() {
	if len(mm.monitors) == 0 {
		mm.loggBook.EnterLogAndPrint("No monitors found.", logger.LogTypes.Error, nil)
	}

	for _, monitor := range mm.monitors {
		monitor.mu.Lock()
		if monitor.DisplayType == Invalid {
			monitor.mu.Unlock()
			continue
		}

		var command string
		switch monitor.DisplayType {
		case AppleDisplay:
			// current brightness...
			if mm.userConfig.Experimental.UseSerialIDForASD {
				command = fmt.Sprintf("asdbctl --serial %s get", monitor.SerialNum)
			} else {
				command = "asdbctl get"
			}

			out, err := cmds.ExecCommand(command, false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}
			monitor.currentBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}

			// max brightness...
			if mm.userConfig.Experimental.UseSerialIDForASD {
				command = fmt.Sprintf("asdbctl --serial %s max", monitor.SerialNum)
			} else {
				command = "asdbctl max"
			}

			out, err = cmds.ExecCommand(command, false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}
			monitor.maxBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
			}

		case DDC:
			command = fmt.Sprintf("ddcutil getvcp 10 --bus %s", monitor.BusNum)
			out, err := cmds.ExecCommand(command, false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}

			parts := strings.Split(out, ":")
			if len(parts) != 2 || len(parts[1]) == 0 {
				continue
			}

			for str := range strings.SplitSeq(parts[1], ",") {
				strs := strings.Split(str, "=")
				if len(strs) != 2 || len(strs[1]) == 0 {
					continue
				}
				trimmed := strings.TrimSpace(strs[0])

				if strings.HasPrefix(trimmed, "current") {
					monitor.currentBrightness, err = strconv.ParseFloat(strs[1], 64)
					if err != nil {
						monitor.DisplayType = Invalid
						mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
						break
					}
					continue
				}
				if strings.HasPrefix(trimmed, "max") {
					monitor.maxBrightness, err = strconv.ParseFloat(strs[1], 64)
					if err != nil {
						monitor.DisplayType = Invalid
						mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
						break
					}
				}
			}

			if (monitor.maxBrightness <= 0) || (monitor.currentBrightness < 0) {
				monitor.DisplayType = Invalid
				err := errors.New("Invalid output from the following command - " + command + "\nThe output is - " + out)
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
			}

		case Backlight:
			// max brightness value...
			out, err := cmds.ExecCommand("brightnessctl m", false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}
			monitor.maxBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}

			// current brightness value...
			out, err = cmds.ExecCommand("brightnessctl g", false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
				break
			}
			monitor.currentBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Warning, nil)
			}
		}
		monitor.mu.Unlock()
	}
}
