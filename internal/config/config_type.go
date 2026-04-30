package config

// Overall Config Struct
type Config struct {
	Logs Logs `toml:"Logs"`
}

// General Section

// Log Section
type Logs struct {
	Level           int    `toml:"Level"`
	Panic           bool   `toml:"Panic"`
	DirectoryPath   string `toml:"DirectoryPath"`
	SaveLogsOnError bool   `toml:"SaveLogsOnError"`
}
