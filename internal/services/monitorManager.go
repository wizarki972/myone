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
	"strings"
	"sync"
	"time"

	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/logger"
	"github.com/wizarki972/myone/internal/utils/pkg"
)

// todos
// dropping temporary/used variable from memory
// dependency warning in New(), maybe notifications too.
// checking pending -- brightness changing
// hyprland IPC monitoring && srevice runner
// communication other services
// gaurd rails for low brightness

const HYPRCTL_MONITORS_CMD = "hyprctl -j monitors"

type DisplayType int

const (
	Backlight DisplayType = iota
	DDC
	AppleDisplay
	Invalid
)

type hyprMonitor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Make        string `json:"make"`
}

type Monitor struct {
	// static
	infoLock    bool
	cardName    string
	BusNum      string
	SerialNum   string
	Backlight   string
	DisplayType DisplayType

	// non-static
	Brigntness float64
	lastUpdate time.Time
	mu         sync.Mutex
}

func (monitor *Monitor) setBrightness(value float64, userConfig *config.Config, loggBook *logger.LogBook) {
	var command string
	switch monitor.DisplayType {
	case Backlight:
		if len(monitor.Backlight) == 0 {
			loggBook.EnterLogAndPrint(fmt.Sprintf("For display %s no backlight was found, but the display type is backlight.", monitor.cardName), logger.LogTypes.Warning, nil)
		} else {
			command = fmt.Sprintf("brightnessctl --device %s s %.0f%%", monitor.Backlight, value)
			break
		}
		fallthrough
	case DDC:
		if len(monitor.BusNum) == 0 {
			loggBook.EnterLogAndPrint(fmt.Sprintf("For display %s no bus number was found, but the display type is DDC ", monitor.cardName), logger.LogTypes.Warning, nil)
		} else {
			command = fmt.Sprintf("ddcutil -b %s setvcp 10 %.0f", monitor.BusNum, value)
			break
		}
		fallthrough
	case AppleDisplay:
		if len(monitor.SerialNum) == 0 && userConfig.Experimental.UseSerialIDForASD {
			loggBook.EnterLogAndPrint(fmt.Sprintf("For display %s no bus number was found, but the display type is Apple Studio Display ", monitor.cardName), logger.LogTypes.Warning, nil)
		} else {
			if userConfig.Experimental.UseSerialIDForASD {
				command = fmt.Sprintf("asdbctl -s %s set %.0f", monitor.SerialNum, value)
			} else {
				command = fmt.Sprintf("asdbctl set %.0f", value)
			}
			break
		}
		fallthrough
	default:
		loggBook.EnterLogAndPrint("Performing default action", logger.LogTypes.Warning, nil)
		command = fmt.Sprintf("brightnessctl s %.0f%%", value)
	}

	if len(command) == 0 {
		loggBook.EnterLogAndPrint("Something thats not supposed to happen happened, since no command was chosen, the service cannot change the brightness.", logger.LogTypes.Warning, nil)
	}

	if _, err := cmds.ExecCommand(command, false, false); err != nil {
		loggBook.EnterLogAndPrint("Error in executing this command --> "+command, logger.LogTypes.Warning, err)
	}

	loggBook.EnterLogAndPrint(fmt.Sprintf("%s --> brightness changed from %.2f to %.2f", monitor.cardName, monitor.Brigntness, value), logger.LogTypes.Info, nil)
}

type MonitorManagaer struct {
	loggBook        *logger.LogBook
	userConfig      *config.Config
	hyprlandSocket2 string
	mu              sync.RWMutex
	ddcutilPresent  bool
	asdbctlPresent  bool

	monitors map[string]*Monitor

	increment float64
}

func NewMonitorManager(loggBook *logger.LogBook, userConfig *config.Config) *MonitorManagaer {
	if userConfig.Experimental.UseSerialIDForASD {
		loggBook.EnterLogAndPrint("Experimental :: Using serial ID of Monitors. Helps with multiple Apple Studio Displays", logger.LogTypes.Info, nil)
	}

	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	hyprlandInstanceSign := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if len(strings.TrimSpace(runtimeDir)) == 0 {
		loggBook.EnterLogAndPrint("Cannot get XDG Runtime Directory environment variable.", logger.LogTypes.Error, errors.New("cannot get XDG Runtime Directory environment variable"))
	}
	if len(strings.TrimSpace(hyprlandInstanceSign)) == 0 {
		loggBook.EnterLogAndPrint("Cannot get Hyprland Instance Signature environment variable.", logger.LogTypes.Error, errors.New("cannot get Hyprland Instance Signature environment variable"))
	}

	return &MonitorManagaer{
		userConfig:      userConfig,
		loggBook:        loggBook,
		hyprlandSocket2: filepath.Join(runtimeDir, "hypr", hyprlandInstanceSign, ".socket2.sock"),
		increment:       0.05,
		ddcutilPresent:  pkg.IsPkgInstalled("ddcutil"),
		asdbctlPresent:  pkg.IsPkgInstalled("asdbctl"),
	}
}

func (mm *MonitorManagaer) HyprlandIPCListener() {
	conn, err := net.Dial("unix", mm.hyprlandSocket2)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Cannot dial hyprland socket2 - "+mm.hyprlandSocket2, logger.LogTypes.Error, err)
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

func (mm *MonitorManagaer) Discover() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mm.monitors = make(map[string]*Monitor)
	mm.getMonitors()

	mm.parseDDCUTIL()

}

func (mm *MonitorManagaer) getMonitors() {
	// brightnessctl i output...
	bctlRegExp := regexp.MustCompile(`Device '([^']+)' of class 'backlight'`)
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

	// getting compositor recognized monitors...
	output, err := cmds.ExecCommandBytes(HYPRCTL_MONITORS_CMD, true)
	if err != nil {
		mm.loggBook.EnterLogAndPrint("Error while executing command - "+HYPRCTL_MONITORS_CMD, logger.LogTypes.Error, err)
	}
	tempMonitors := make([]hyprMonitor, 0)
	if err := json.Unmarshal(output, &tempMonitors); err != nil {
		mm.loggBook.EnterLogAndPrint("Failed to parse json values from hyprctl "+HYPRCTL_MONITORS_CMD, logger.LogTypes.Error, err)
	}

	// processing reorganized...
	for _, monitor := range tempMonitors {
		mm.monitors[monitor.Name] = &Monitor{
			infoLock: false,
			cardName: monitor.Name,
		}

		// drm matching...
		globPath := filepath.Join("/sys/class/drm", "*"+monitor.Name+"*")
		matches, err := filepath.Glob(globPath)
		if err != nil {
			mm.loggBook.EnterLogAndPrint("Cannot get drm entry for "+monitor.Name, logger.LogTypes.Warning, nil)
			mm.monitors[monitor.Name].DisplayType = Invalid
			continue
		}

		// even if more than one match is found, the first one is used...
		drmInfo, err := os.Stat(matches[0])
		if err != nil {
			mm.loggBook.EnterLogAndPrint("Cannot get drm entry for "+monitor.Name, logger.LogTypes.Warning, nil)
			mm.monitors[monitor.Name].DisplayType = Invalid
			continue
		}

		// eDP, backlight check....
		if strings.Contains(monitor.Name, "eDP") && len(bctlDevices) > 0 {
			// backlight matching...
			for _, backlightName := range bctlDevices {
				backlightPath := filepath.Join("/sys/class/drm", drmInfo.Name(), backlightName)
				_, err := os.Stat(backlightPath)
				if err != nil {
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
				}
				mm.monitors[monitor.Name].DisplayType = Backlight
				mm.monitors[monitor.Name].Backlight = backlightName
				mm.monitors[monitor.Name].infoLock = true
				break
			}

			if mm.monitors[monitor.Name].infoLock {
				continue
			}
		}

		// Apple Check...
		preprocessedDescription := strings.ReplaceAll(strings.ToLower(monitor.Description), " ", "")
		if strings.Contains(preprocessedDescription, "apple") || strings.Contains(preprocessedDescription, "studiodisplay") {
			mm.monitors[monitor.Name].DisplayType = AppleDisplay
			continue
		}

		preprocessedMake := strings.ReplaceAll(strings.ToLower(monitor.Make), " ", "")
		if strings.Contains(preprocessedMake, "apple") || strings.Contains(preprocessedMake, "applecomputerinc") {
			mm.monitors[monitor.Name].DisplayType = AppleDisplay
			continue
		}

		// survived those then its a DDC...
		mm.monitors[monitor.Name].DisplayType = DDC
	}
}

func (mm *MonitorManagaer) parseDDCUTIL() {
	out, err := cmds.ExecCommand("ddcutil detect", false, true)
	if err != nil {
		if !mm.ddcutilPresent {
			mm.loggBook.EnterLogAndPrint("ddcutil dependency not found", logger.LogTypes.Warning, err)
			return
		}
		mm.loggBook.EnterLogAndPrint(err.Error(), logger.LogTypes.Error, err)
	}

	i2cBusMatch := regexp.MustCompile(`/dev/i2c-(\d+)`)
	connectorMatch := regexp.MustCompile(`DRM_connector:\s+([\w-]+)`)
	serialMatch := regexp.MustCompile(`Serial number:\s+([\w-]+)`)

	for block := range strings.SplitSeq(out, "\n\n") {
		// skipping failure messages
		if strings.Contains(block, "Failed") {
			var fildtered []string
			for line := range strings.SplitSeq(block, "\n") {
				if !strings.HasPrefix(strings.TrimSpace(line), "Failed") {
					fildtered = append(fildtered, line)
				}
			}
			block = strings.Join(fildtered, "\n")
		}

		// Invalid display check
		if strings.Contains(block, "Invalid display") {
			continue
		}

		// getting card name
		matches := connectorMatch.FindStringSubmatch(block)
		if len(matches) <= 1 {
			continue
		}
		matches = strings.SplitN(matches[1], "-", 2)
		if len(matches) <= 1 {
			continue
		}
		// getting monitor...
		monitor, ok := mm.monitors[matches[1]]
		if !ok {
			mm.loggBook.EnterLogAndPrint("Skipped "+matches[1]+", this monitor is not recognized by the compositor.", logger.LogTypes.Warning, err)
			continue
		}
		// separarte because logging is required above...
		if monitor.infoLock {
			continue
		}

		// bus number
		i2cbusMatches := i2cBusMatch.FindStringSubmatch(block)
		if len(i2cbusMatches) > 1 {
			monitor.BusNum = i2cbusMatches[1]
		} else if monitor.DisplayType == DDC {
			// sine there are use for an serial ID in ddc when no bus number is found, skipping it...
			monitor.infoLock = true
			monitor.DisplayType = Invalid
			continue
		}

		// serial number
		if mm.userConfig.Experimental.UseSerialIDForASD {
			serialMatches := serialMatch.FindStringSubmatch(block)
			if len(serialMatches) > 1 {
				monitor.SerialNum = serialMatches[1]
			}
		}

		monitor.infoLock = true
	}
}
