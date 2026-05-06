package config

// Overall Config Struct
type Config struct {
	Battery      Battery      `toml:"Battery"`
	Logs         Logs         `toml:"Logs"`
	Experimental Experimental `toml:"Experimental"`
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

// Experimental
type Experimental struct {
	UseSerialIDForASD bool `toml:"UseSerialIDForASD"`
}
