package avistaz

import (
	"github.com/BurntSushi/toml"
)



type Config struct {
	Avistaz_session string
	XSRF_TOKEN string
	Avistazlove string
	Cfduid string
	Remember_id string
	Remember string
}

func NewConfig() *Config {
	c := Config{}
	c.initialize()
	return &c
}

func (c *Config) initialize() {
}

func (c *Config) Load(configfile string) error {
	_, err := toml.DecodeFile(configfile, c)
	return err
}

func (c *Config) String() string {
	return c.Avistaz_session
}
