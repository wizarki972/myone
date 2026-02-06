package display

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wizarki972/myone/internal/utils/cmds"
	"github.com/wizarki972/myone/internal/utils/config"
)

var hyprlandMonitorsComamnd = "hyprctl -j monitors"

type Monitor struct {
	Name           string
	Card_name      string
	Backlight_name string
}

type MonitorDaemon struct {
	Sock           string
	HyprlandSock   string
	Events         chan string
	Monitors       map[string]*Monitor
	default_config config.Config

	mu sync.RWMutex
}

func NewMonitorDaemon() *MonitorDaemon {
	runtime := os.Getenv("XDG_RUNTIME_DIR")
	hypr_sign := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if runtime == "" || hypr_sign == "" {
		panic(errors.New("cannot get XDG_RUNTIME_DIR environment variable"))
	}

	md := &MonitorDaemon{
		Sock:           filepath.Join(runtime, "myone-display-monitor.sock"),
		HyprlandSock:   filepath.Join(runtime, "hypr", hypr_sign, ".socket2.sock"),
		Events:         make(chan string),
		Monitors:       make(map[string]*Monitor),
		default_config: config.Default_Config,
	}

	var monitors []hyprMonitor
	command := "hyprctl -j monitors"
	data := cmds.ExecCommandWithOutput(command)
	if err := json.Unmarshal(data, &monitors); err != nil {
		panic(err)
	}
	for _, monitor := range monitors {
		md.GenerateMonitor(monitor.Name)
	}

	return md
}

func (md *MonitorDaemon) GenerateMonitor(name string) {
	monitor := &Monitor{
		Name:      name,
		Card_name: "",
	}

	// getting drm card name
	drm_entries, err := os.ReadDir("/sys/class/drm") // name
	if err != nil {
		panic(err)
	}
	for _, entry := range drm_entries {
		if strings.HasSuffix(entry.Name(), monitor.Name) {
			monitor.Card_name = entry.Name()
		}
	}
	if monitor.Card_name == "" {
		panic(errors.New("monitor drm entry not found"))
	}

	// getting all available backlight entries
	backlight_entries, err := os.ReadDir("/sys/class/backlight")
	if err != nil {
		panic(err)
	}

	// finding the correct backlight entry for the monitor
	for _, backlight_entry := range backlight_entries {
		base := filepath.Join("/sys/class/drm", monitor.Card_name, backlight_entry.Name())
		info, err := os.Stat(base)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			panic(err)
		}
		if info.IsDir() { //checking whether the info is a symlink
			monitor.Backlight_name = backlight_entry.Name()
			break
		}
	}

	// adding the monitor
	md.mu.Lock()
	defer md.mu.Unlock()
	md.Monitors[monitor.Name] = monitor
}

func (md *MonitorDaemon) ChangeBrightness(name, value string) {
	command := fmt.Sprintf("brightnessctl -d %s set %s", md.Monitors[name].Backlight_name, value)
	err := cmds.ExecComamndWithError(command)
	if err == nil {
		md.mu.RLock()
		defer md.mu.RUnlock()
		swayOSDNotify(md.Monitors[name].Backlight_name)
	} else {
		panic(err)
	}
}

func (md *MonitorDaemon) HyprlandIPCListenr() {
	conn, err := net.Dial("unix", md.HyprlandSock)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	md.Events <- "Monitoring to Hyprland IPC scoket..."
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "monitorremoved") || strings.HasPrefix(line, "monitoradded") {
			name := strings.Split(line, ">>")[1]
			md.Events <- "New Monitor Detected"
			md.GenerateMonitor(name)
		}
	}

}

func (md *MonitorDaemon) StartListener() {
	os.Remove(md.Sock)

	listener, err := net.Listen("unix", md.Sock)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	md.Events <- "Listening for brightness control in myone socket..."
	for {
		c, _ := listener.Accept()

		go func(conn net.Conn) {
			defer conn.Close()
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				args := strings.TrimSpace(scanner.Text())
				md.Events <- "Got argument to change brightness to " + args
				name, err := activeMonitor()
				if err != nil {
					panic(err)
				}

				md.ChangeBrightness(name, args)
			}
		}(c)

	}
}

func (md *MonitorDaemon) StartDaemon() {
	go md.HyprlandIPCListenr()
	go md.StartListener()

	for {
		slog.Info(<-md.Events)
	}
}
