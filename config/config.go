package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

const DummyMode = "dummy"
const DbMode = "db"

type Config struct {
	Mode  string `yaml:"mode"`
	Dsn   string `yaml:"dsn"`
	Redis string `yaml:"redis"`
}

func NewConfig(filePath string) (*Config, error) {
	var ans Config
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(b, &ans); err != nil {
		return nil, err
	}
	return &ans, nil
}
