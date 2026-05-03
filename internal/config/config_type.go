package config

// Overall Config Struct
type Config struct {
	Battery Battery `toml:"Battery"`
	Logs    Logs    `toml:"Logs"`
}

// General Section

// Battery
type Battery struct {
	Threshold int `toml:"Threshold"`
}

// Logging
type Logs struct {
	Level           int    `toml:"Level"`
	Panic           bool   `toml:"Panic"`
	DirectoryPath   string `toml:"DirectoryPath"`
	SaveLogsOnError bool   `toml:"SaveLogsOnError"`
}
