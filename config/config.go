package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Dsn  string `yaml:"dsn"`
	Port string `yaml:"port"`
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
