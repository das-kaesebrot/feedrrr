package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type FeedrrrConfig struct {
	Jobs  map[string]JobConfig `yaml:"jobs"`
	Sinks map[string][]string  `yaml:"sinks"`
}

type JobConfig struct {
	Sinks    []string `yaml:"sinks"`
	Schedule string   `yaml:"schedule"`
	Source   string   `yaml:"source"`
}

func ParseConfig(appName string) (*FeedrrrConfig, error) {
	appName = strings.ToValidUTF8(strings.ToLower(appName), "")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("/etc/%v/", appName))

	// respect XDG base directory specification https://specifications.freedesktop.org/basedir/latest/#variables
	viper.AddConfigPath(fmt.Sprintf("$XDG_CONFIG_HOME/%v/", appName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%v/", appName))
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &FeedrrrConfig{}
	err = viper.UnmarshalExact(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
