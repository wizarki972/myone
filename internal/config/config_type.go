package config

// Overall Config Struct
type Config struct {
	Battery      Battery      `toml:"battery"`
	Monitor      Monitor      `toml:"monitor"`
	Logs         Logs         `toml:"logs"`
	Experimental Experimental `toml:"experimental"`
}

// General Section

// Battery
type Battery struct {
	Threshold int `toml:"threshold"`
}

type Monitor struct {
	Check_Backlight_For_ALL_Displays bool `toml:"check_backlight_for_all_displays"`
	Ignore_Apple_Studio_Displays     bool `toml:"ignore_apple_studio_displays"`
}

// Logging
type Logs struct {
	Level              int    `toml:"level"`
	Panic              bool   `toml:"panic_on_error"`
	Directory_Path     string `toml:"directory_path"`
	Save_Logs_On_Error bool   `toml:"save_logs_on_error"`
	Logs_Save_Interval int    `toml:"log_save_interval"`
}

// Experimental
type Experimental struct {
	Use_Serial_ID_For_Apple_Studio_Displays bool `toml:"use_serial_id_for_apple_studio_displays"`
}
