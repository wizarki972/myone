package config

// Overall Config Struct
type Config struct {
	general general `toml:"General"`
	Logs    Logs    `toml:"Logs"`
}

// General Section
type general struct {
	configPath string
}

// Log Section
type Logs struct {
	Level           int    `toml:"Level"`
	Panic           bool   `toml:"Panic"`
	DirectoryPath   string `toml:"DirectoryPath"`
	SaveLogsOnError bool   `toml:"SaveLogsOnError"`
}
