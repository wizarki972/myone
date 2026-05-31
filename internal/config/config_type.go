package config

// Overall Config Struct
type Config struct {
	Battery      Battery      `toml:"battery"`
	Logs         Logs         `toml:"logs"`
	Experimental Experimental `toml:"experimental"`
}

// General Section

// Battery
type Battery struct {
	Threshold int `toml:"threshold"`
}

// Logging
type Logs struct {
	Level           int    `toml:"level"`
	Panic           bool   `toml:"panic"`
	DirectoryPath   string `toml:"directory_path"`
	SaveLogsOnError bool   `toml:"save_logs_on_error"`
	LogSaveInterval int    `toml:"log_save_interval"`
}

// Experimental
type Experimental struct {
	UseSerialIDForASD bool `toml:"use_serial_id_for_apple_studio_displays"`
}
