package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type FeedrrrConfig struct {
	Jobs  map[string]JobConfig `mapstructure:"jobs"`
	Sinks map[string][]string  `mapstructure:"sinks"`
}

type JobConfig struct {
	Sinks        []string `mapstructure:"sinks"`
	Schedule     string   `mapstructure:"schedule"`
	Source       string   `mapstructure:"source"`
	UsePlainText bool     `mapstructure:"plaintext,omitempty"`
	Prefix       string   `mapstructure:"prefix,omitempty"` // title prefix
}

func ParseConfig(appName string, configFileOverride string) (*FeedrrrConfig, error) {
	appName = strings.ToValidUTF8(strings.ToLower(appName), "")

	if configFileOverride != "" {
		viper.SetConfigFile(configFileOverride)
		slog.Info("Using explicit config file path", "configFileOverride", configFileOverride)
	} else {
		setupDefaultConfigPaths(appName)
		slog.Info("Using default config search paths")
	}

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

func setupDefaultConfigPaths(appName string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(fmt.Sprintf("/etc/%v/", appName))

	// respect XDG base directory specification https://specifications.freedesktop.org/basedir/latest/#variables
	viper.AddConfigPath(fmt.Sprintf("$XDG_CONFIG_HOME/%v/", appName))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%v/", appName))
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()
}
