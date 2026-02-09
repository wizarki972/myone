package config

type Config struct {
	general General `toml:"general"`
	Display Display `toml:"display"`
}

type General struct {
	config_path string
}

type Display struct {
	Backlight_name string `toml:"backlight_name"`
}
