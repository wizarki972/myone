package services

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wizarki972/myone/internal/common"
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
	drmConnectorMatch = regexp.MustCompile(`DRM_connector:\s*((card\d+)-([\w-]+))`)
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
	minBrightness     float64
	currentBrightness float64
	mu                sync.RWMutex
}

func (monitor *Monitor) setBrightness(value float64, userConfig *config.Config, loggBook *logger.LogBook) {
	command := fmt.Sprintf("brightnessctl s %.0f", value)

	switch monitor.DisplayType {
	case Backlight:
		if len(monitor.Backlight) > 0 {
			command = fmt.Sprintf("brightnessctl --device %s s %.0f", monitor.Backlight, value)
			break
		}
		err := fmt.Errorf("For display %s no backlight was found, but the display type is backlight.", monitor.Name)
		loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	case DDC:
		if len(monitor.BusNum) > 0 {
			command = fmt.Sprintf("ddcutil -b %s setvcp 10 %.0f", monitor.BusNum, value)
			break
		}
		err := fmt.Errorf("For display %s no bus number was found, but the display type is DDC ", monitor.Name)
		loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	case AppleDisplay:
		command = fmt.Sprintf("asdbctl set %.0f", value)

		if !userConfig.Experimental.Use_Serial_ID_For_Apple_Studio_Displays {
			break
		}
		if len(monitor.SerialNum) > 0 {
			command = fmt.Sprintf("asdbctl -s %s set %.0f", monitor.SerialNum, value)
			break
		}
		err := fmt.Errorf("For display %s no serial number was found, but the display type is Apple Studio Display ", monitor.Name)
		loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	default:
		loggBook.EnterLogAndPrint("Invalid monitor type.", logger.LogTypes.Error, errors.New("invalid monitor type"))
		return
	}

	if _, err := cmds.ExecCommand(command, false, false); err != nil {
		loggBook.EnterLogAndPrint("Error in executing this command --> "+command, logger.LogTypes.Error, err)
		return
	}
	loggBook.EnterLogAndPrint(fmt.Sprintf("%s --> brightness changed from %.2f to %.2f", monitor.Name, monitor.currentBrightness, value), logger.LogTypes.Info, nil)
	monitor.currentBrightness = value
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
	brightnessChangeRequests chan MMCommand
	ctx                      context.Context
}

func NewMonitorManager(loggBook *logger.LogBook, userConfig *config.Config) *MonitorManager {
	if userConfig.Experimental.Use_Serial_ID_For_Apple_Studio_Displays {
		loggBook.EnterLogAndPrint("Experimental :: Using serial ID of Monitors. Helps with multiple Apple Studio Displays", logger.LogTypes.Info, nil)
	}

	runtimeDir, err := fldir.GetRuntimeDir()
	if err != nil {
		loggBook.EnterLogAndPrint("Failed to get a proper runtime directory.", logger.LogTypes.Error, err)
		return nil
	}
	hyprlandInstanceSign := strings.TrimSpace(os.Getenv("HYPRLAND_INSTANCE_SIGNATURE"))
	if len(runtimeDir) == 0 {
		loggBook.EnterLogAndPrint("Cannot get XDG Runtime Directory environment variable.", logger.LogTypes.Error, errors.New("cannot get XDG Runtime Directory environment variable"))
		return nil
	}
	if len(hyprlandInstanceSign) == 0 {
		loggBook.EnterLogAndPrint("Cannot get Hyprland Instance Signature environment variable.", logger.LogTypes.Error, errors.New("cannot get Hyprland Instance Signature environment variable"))
		return nil
	}

	mm := &MonitorManager{
		userConfig: userConfig,
		loggBook:   loggBook,

		mmSocket:        filepath.Join(runtimeDir, "myone", "monitor.sock"),
		hyprlandSocket2: filepath.Join(runtimeDir, "hypr", hyprlandInstanceSign, ".socket2.sock"),

		ddcutilPresent: pkg.IsPkgInstalled("ddcutil"),
		asdbctlPresent: pkg.IsPkgInstalled("asdbctl"),

		monitors:                 make(map[string]*Monitor),
		brightnessChangeRequests: make(chan MMCommand, 10),
		ctx:                      nil,
	}

	if !mm.ddcutilPresent {
		mm.loggBook.EnterLogAndPrint("Missing dependency - ddcutil.", logger.LogTypes.Error, errors.New("Missing dependency - ddcutil."))
		return nil
	}

	if !mm.asdbctlPresent && !userConfig.Monitor.Ignore_Apple_Studio_Displays {
		mm.loggBook.EnterLogAndPrint("Missing dependency - ddcutil.", logger.LogTypes.Error, errors.New("Missing dependency - asdbctl."))
		return nil
	}

	return mm
}

// BELOW CODE FOR - BRIGHTNESS REQUESTS

// handles new  brightness change requets
func (mm *MonitorManager) brightnessRequestHandler() {
	for {
		select {
		case mmCommand := <-mm.brightnessChangeRequests:
			// Err checks....
			if len(mmCommand.TargetMonitor) == 0 || mmCommand.Value <= 0 {
				mm.loggBook.EnterLogAndPrint("Invalid brightness change values received.", logger.LogTypes.Error, errors.New("invalid brightness change values received"))
				return
			}

			// monitor fetching
			mm.mu.RLock()
			monitor, ok := mm.monitors[mmCommand.TargetMonitor]
			if !ok {
				mm.loggBook.EnterLogAndPrint("Invalid monitor name received - "+mmCommand.TargetMonitor, logger.LogTypes.Error, errors.New("invalid monitor name received - "+mmCommand.TargetMonitor))
				return
			}

			// get monitor values....
			monitor.mu.Lock()
			if monitor.DisplayType == Invalid {
				monitor.mu.Unlock()
				mm.mu.RUnlock()
				mm.loggBook.EnterLogAndPrint("Invalid display type.", logger.LogTypes.Error, errors.New("invalid display type"))
				return
			}
			currentBrightness := monitor.currentBrightness
			maxBrightness := monitor.maxBrightness
			minBrightness := monitor.minBrightness

			// brightness value calc-ing...
			mmCommand.Value = maxBrightness * (mmCommand.Value / 100)
			switch mmCommand.Prefix {
			case '+':
				mmCommand.Value = max(min(currentBrightness+mmCommand.Value, maxBrightness), minBrightness)
			case '-':
				mmCommand.Value = max(min(currentBrightness-mmCommand.Value, maxBrightness), minBrightness)
			}

			// setting brightness values...
			monitor.setBrightness(mmCommand.Value, mm.userConfig, mm.loggBook)
			monitor.mu.Unlock()
			mm.mu.RUnlock()
		case <-mm.ctx.Done():
			return
		}
	}
}

// BELOW CODE FOR - SOCKET LISTENERS & SERVICE STARTER

func (mm *MonitorManager) hyprlandIPCListener() {
	mm.loggBook.EnterLogAndPrint("Starting Hyprland IPC listener...", logger.LogTypes.Info, nil)
	conn, err := net.Dial("unix", mm.hyprlandSocket2)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Cannot dial hyprland socket2 - "+mm.hyprlandSocket2, logger.LogTypes.Error, err)
		return
	}
	defer conn.Close()

	lines := make(chan string)
	connScanner := bufio.NewScanner(conn)

	go func() {
		defer close(lines)
		for connScanner.Scan() {
			lines <- connScanner.Text()
		}
	}()
	mm.loggBook.EnterLogAndPrint("Monitoring Hyprland IPC socket -> "+mm.hyprlandSocket2, logger.LogTypes.Info, nil)

	var dCtx context.Context = nil
	var dCancel context.CancelFunc = nil

	for {
		select {
		case <-mm.ctx.Done():
			if dCancel != nil {
				dCancel()
			}
			return
		case line, ok := <-lines:
			if !ok {
				mm.loggBook.EnterLogAndPrint("Hyprland socket connection closed.", logger.LogTypes.Error, connScanner.Err())
				if dCancel != nil {
					dCancel()
				}
				return
			}

			if dCancel != nil {
				dCancel()
				dCancel = nil
			}

			dCtx, dCancel = context.WithCancel(mm.ctx)
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "monitorremoved") || strings.HasPrefix(line, "monitoradded") {
				mm.loggBook.EnterLogAndPrint("Change detected in the number of monitors. Updating monitors information.", logger.LogTypes.Info, nil)
				go func(Ctx context.Context, cancel context.CancelFunc) {
					defer cancel()
					mm.Discover(Ctx)
				}(dCtx, dCancel)
			}

		}
	}
}

func (mm *MonitorManager) StartService() {
	isRunning, pid, err := isOldProcessRunning(common.MONITOR_MON_PID_FILE_NAME)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Cannot determine whether an old process is running or not.", logger.LogTypes.Warning, nil)
	}

	if isRunning {
		if err = killProcess(pid); err != nil {
			mm.loggBook.EnterLogAndPrint("Failed to kill an old process.", logger.LogTypes.Error, errors.New("Failed to kill an old process."))
			return
		}
	}

	if err := savePID(common.MONITOR_MON_PID_FILE_NAME, os.Getpid()); err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to save PID  so the current service is stopped.", logger.LogTypes.Error, err)
		return
	}

	// main...
	var cancel context.CancelFunc = nil
	mm.ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup

	if len(mm.monitors) == 0 {
		mm.Discover(mm.ctx)
	}

	// starting auto logs saver...
	wg.Go(func() {
		defer cancel()
		mm.loggBook.StartAutoLogSaver(mm.ctx)
	})

	// Hyprland IPC listener
	wg.Go(func() {
		defer cancel()
		mm.hyprlandIPCListener()
	})

	// Brightness change requests handler...
	wg.Go(func() {
		defer cancel()
		mm.brightnessRequestHandler()
	})

	// listens for client requests...
	wg.Go(func() {
		defer cancel()
		mm.requestListener()
	})

	wg.Wait()
}

// listens for all kinds of monitor releated requests from the socket, but currently only for brightness change requests.
func (mm *MonitorManager) requestListener() {
	// path checks...
	os.Remove(mm.mmSocket)
	if err := fldir.CreateDirectory(filepath.Dir(mm.mmSocket)); err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to create directory for the listening socket - "+mm.mmSocket, logger.LogTypes.Error, err)
		return
	}

	// create listener...
	listener, err := net.Listen("unix", mm.mmSocket)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to listen from socket - "+mm.mmSocket, logger.LogTypes.Error, err)
		return
	}
	defer listener.Close()

	// local context...
	localCtx, localCancel := context.WithCancel(mm.ctx)
	defer localCancel()

	// ctx done listener...
	go func() {
		<-localCtx.Done()
		if mm.ctx.Err() != nil {
			mm.loggBook.EnterLogAndPrint("Context cancelled. closing client socket listener.", logger.LogTypes.Error, err)
		}
		listener.Close()
	}()

	// start lsitening...
	mm.loggBook.EnterLogAndPrint("Listening for requets...", logger.LogTypes.Info, nil)
	for {
		c, err := listener.Accept()
		if err != nil {
			select {
			case <-mm.ctx.Done():
				return
			default:
				if localCtx.Err() != nil {
					return
				}
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				continue
			}
		}

		// spin off each connection into their own concurrent process...
		go func(conn net.Conn) {
			defer conn.Close()

			decoder := json.NewDecoder(conn)
			for {
				// close any connection older than 5 seconds after receiving a request...
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				var request MMCommand
				if err := decoder.Decode(&request); err != nil {
					if err == io.EOF {
						return
					}
					mm.loggBook.EnterLogAndPrint("Failed to decode client request.", logger.LogTypes.Error, err)
					return
				}

				mm.loggBook.EnterLogAndPrint(fmt.Sprintf("Received request ==> %s, Value: %.2f,Target Monitor: %s,Prefix: %c", request.Command, request.Value, request.TargetMonitor, request.Prefix), logger.LogTypes.Info, nil)
				if len(request.Command) == 0 || request.Value <= 0 || len(request.TargetMonitor) == 0 {
					mm.loggBook.EnterLogAndPrint("Request is invalid", logger.LogTypes.Error, errors.New("invalid request"))
					// instead of breaking of because of an error, we continue the loop to check for new valid requests...
					continue
				}

				switch request.Command {
				case "brightness":
					mm.brightnessChangeRequests <- request
				default:
					mm.loggBook.EnterLogAndPrint("Invalid command type ==> "+request.Command, logger.LogTypes.Error, errors.New("invalid command type ==> "+request.Command))
				}
			}
		}(c)
	}
}

// BELOW CODE IS FOR - DISCOVER MONITORS AND GET BRIGHTNESS VALUES

// Discover all available monitors
func (mm *MonitorManager) Discover(dCtx context.Context) {
	if err := dCtx.Err(); err != nil {
		mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		return
	}

	mm.loggBook.EnterLogAndPrint("Scanning for all monitors...", logger.LogTypes.Info, nil)
	tempMonitorsSlice := make(map[string]*Monitor)
	if err := mm.prepareMonitorsSlice(dCtx, tempMonitorsSlice); err != nil {
		return
	}

	if err := mm.fillMonitorValues(dCtx, tempMonitorsSlice); err != nil {
		return
	}

	if err := mm.getBrightnessValues(dCtx, tempMonitorsSlice); err != nil {
		return
	}

	if len(tempMonitorsSlice) > 0 {
		mm.mu.Lock()
		mm.monitors = tempMonitorsSlice
		mm.mu.Unlock()
		return
	}
	mm.loggBook.EnterLogAndPrint("No monitor/display devices are found", logger.LogTypes.Warning, nil)
}

// prepare monitor slice with all compositor recognized monitors...
func (mm *MonitorManager) prepareMonitorsSlice(dCtx context.Context, tempMonitorSlice map[string]*Monitor) error {
	// getting compositor recognized monitors...
	output, err := cmds.ExecCommandContextBytes(dCtx, HYPRCTL_MONITORS_CMD, false, true)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Error while executing command - "+HYPRCTL_MONITORS_CMD, logger.LogTypes.Error, err)
		return err
	}
	tempMonitors := make([]hyprMonitor, 0)
	if err := json.Unmarshal(output, &tempMonitors); err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to parse json values from hyprctl "+HYPRCTL_MONITORS_CMD, logger.LogTypes.Error, err)
		return err
	}

	// preparing monitors slice...
	for _, monitor := range tempMonitors {
		tempMonitorSlice[monitor.Name] = &Monitor{
			Name:              monitor.Name,
			DisplayType:       Unknown,
			maxBrightness:     -1,
			currentBrightness: -1,
		}

		tempMonitorSlice[monitor.Name].mu.Lock()
		// Apple Check...
		preprocessedDescription := strings.ReplaceAll(strings.ToLower(monitor.Description), " ", "")
		if strings.Contains(preprocessedDescription, "apple") || strings.Contains(preprocessedDescription, "studiodisplay") {
			tempMonitorSlice[monitor.Name].DisplayType = AppleDisplay
		} else {
			preprocessedMake := strings.ReplaceAll(strings.ToLower(monitor.Make), " ", "")
			if strings.Contains(preprocessedMake, "apple") || strings.Contains(preprocessedMake, "applecomputerinc") {
				tempMonitorSlice[monitor.Name].DisplayType = AppleDisplay
			}
		}
		tempMonitorSlice[monitor.Name].mu.Unlock()
	}
	return nil
}

// fill monitors slice with necessary values...
func (mm *MonitorManager) fillMonitorValues(dCtx context.Context, tempMonitorsSlice map[string]*Monitor) error {
	// brightnessctl i output...
	bctlOut, err := cmds.ExecCommandContext(dCtx, "brightnessctl i", false, true)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Error cannot run command - 'brightnessctl i'", logger.LogTypes.Error, err)
		return err
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
	ddcUtilOut, err := cmds.ExecCommandContext(dCtx, "ddcutil detect", false, true)
	if err != nil {
		if !mm.ddcutilPresent {
			mm.loggBook.EnterLogAndPrint("ddcutil dependency not found", logger.LogTypes.Error, err)
			return err
		}
		mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
		return err
	}

ddcUtilBlockLoop:
	for block := range strings.SplitSeq(ddcUtilOut, "\n\n") {
		// skipping failure messages & invalid blocks...
		var failedFound, invalidFound bool
		var filtered = []string{}
		for line := range strings.SplitSeq(block, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "Failed") {
				failedFound = true
				continue
			}
			if strings.HasPrefix(trimmed, "Invalid display") {
				invalidFound = true
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
		monitor, ok := tempMonitorsSlice[matches[3]]
		if !ok {
			mm.loggBook.EnterLogAndPrint("Skipped "+matches[1]+", this monitor is not recognized by the compositor.", logger.LogTypes.Error, err)
			continue
		}
		monitor.mu.Lock()
		monitor.cardName = matches[1]

		drmPath := filepath.Join("/sys/class/drm", monitor.cardName)
		if fldir.IsPathExist(drmPath) {
			// apple display - serial number
			if monitor.DisplayType == AppleDisplay {
				if mm.userConfig.Monitor.Ignore_Apple_Studio_Displays {
					monitor.DisplayType = Invalid
					monitor.mu.Unlock()
					continue
				}
				if mm.userConfig.Experimental.Use_Serial_ID_For_Apple_Studio_Displays {
					serialMatches := serialNumberMatch.FindStringSubmatch(block)
					if len(serialMatches) == 2 {
						monitor.SerialNum = serialMatches[1]
					} else {
						monitor.DisplayType = Invalid
						mm.loggBook.EnterLogAndPrint("Failed to get proper serial number for "+monitor.cardName, logger.LogTypes.Error, nil)
					}
				}
				monitor.mu.Unlock()
				continue
			}

			// eDP backlight matching...
			if (mm.userConfig.Monitor.Check_Backlight_For_ALL_Displays || strings.Contains(monitor.Name, "eDP")) && len(bctlDevices) > 0 {
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

			if invalidFound {
				monitor.DisplayType = Invalid
				monitor.mu.Unlock()
				continue
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
	return nil
}

// get the monitors current and max brightness...
func (mm *MonitorManager) getBrightnessValues(dCtx context.Context, tempMonitorsSlice map[string]*Monitor) error {
	if len(tempMonitorsSlice) == 0 {
		err := errors.New("monitors not found")
		mm.loggBook.EnterLogAndPrint("Monitors not found.", logger.LogTypes.Error, err)
		return err
	}

monitorLoop:
	for _, monitor := range tempMonitorsSlice {
		monitor.mu.Lock()
		if monitor.DisplayType == Invalid {
			monitor.mu.Unlock()
			continue monitorLoop
		}

		var command string
		switch monitor.DisplayType {
		case AppleDisplay:
			// current brightness...
			if mm.userConfig.Experimental.Use_Serial_ID_For_Apple_Studio_Displays {
				command = fmt.Sprintf("asdbctl --serial %s get", monitor.SerialNum)
			} else {
				command = "asdbctl get"
			}

			out, err := cmds.ExecCommandContext(dCtx, command, false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.currentBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}

			// max brightness...
			if mm.userConfig.Experimental.Use_Serial_ID_For_Apple_Studio_Displays {
				command = fmt.Sprintf("asdbctl --serial %s max", monitor.SerialNum)
			} else {
				command = "asdbctl max"
			}

			out, err = cmds.ExecCommandContext(dCtx, command, false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.maxBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.minBrightness = monitor.maxBrightness * 0.01

		case DDC:
			command = fmt.Sprintf("ddcutil getvcp 10 --bus %s", monitor.BusNum)
			out, err := cmds.ExecCommandContext(dCtx, command, false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}

			parts := strings.Split(out, ":")
			if len(parts) != 2 || len(parts[1]) == 0 {
				monitor.DisplayType = Invalid
				monitor.mu.Unlock()
				mm.loggBook.EnterLogAndPrint("Failed to get brightness properties of the display - "+monitor.Name, logger.LogTypes.Error, errors.New("failed to get brightness properties of the display - "+monitor.Name))
				continue monitorLoop
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
						mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
						break
					}
					continue
				}
				if strings.HasPrefix(trimmed, "max") {
					monitor.maxBrightness, err = strconv.ParseFloat(strs[1], 64)
					if err != nil {
						monitor.DisplayType = Invalid
						mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
						break
					}
				}
			}

			if (monitor.maxBrightness <= 0) || (monitor.currentBrightness < 0) {
				monitor.DisplayType = Invalid
				err := errors.New("Invalid output from the following command - " + command + "\nThe output is - " + out)
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.minBrightness = monitor.maxBrightness * 0.01

		case Backlight:
			// max brightness value...
			out, err := cmds.ExecCommandContext(dCtx, "brightnessctl m", false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.maxBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}

			// current brightness value...
			out, err = cmds.ExecCommandContext(dCtx, "brightnessctl g", false, true)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.currentBrightness, err = strconv.ParseFloat(out, 64)
			if err != nil {
				monitor.DisplayType = Invalid
				mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				break
			}
			monitor.minBrightness = monitor.maxBrightness * 0.01
		}
		monitor.mu.Unlock()
	}
	return nil
}
