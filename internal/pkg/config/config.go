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
		searchPaths := []string{
			fmt.Sprintf("/etc/%v/", appName),

			// respect XDG base directory specification https://specifications.freedesktop.org/basedir/latest/#variables
			fmt.Sprintf("$XDG_CONFIG_HOME/%v/", appName),
			fmt.Sprintf("$HOME/.config/%v/", appName),
			".",
		}

		setupDefaultConfigPaths(appName, searchPaths)
		slog.Info("Using default config search paths", "searchPaths", searchPaths)
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

func setupDefaultConfigPaths(appName string, searchPaths []string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	for _, searchPath := range searchPaths {
		viper.AddConfigPath(searchPath)
	}

	viper.SetEnvPrefix(appName)
	viper.AutomaticEnv()
}
