package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type ChangeDetectionMode int

const (
	ModePubDate ChangeDetectionMode = iota
	ModeGUID
)

const (
	DefaultChangeDetectionMode        = ModeGUID
	DefaultMessageTemplate     string = `{{.Title}} ({{.PubDate}})

{{ html2text (or .Content .Description "no content") }}

{{.Link}}
`
)

var changeModeMap = map[string]ChangeDetectionMode{
	"pubdate": ModePubDate,
	"guid":    ModeGUID,
}

type FeedrrrConfig struct {
	Jobs  map[string]JobConfig `mapstructure:"jobs"`
	Sinks map[string][]string  `mapstructure:"sinks"`
}

type JobConfig struct {
	Sinks           []string            `mapstructure:"sinks"`
	Schedule        string              `mapstructure:"schedule"`
	Source          string              `mapstructure:"source"`
	UsePlainText    bool                `mapstructure:"plaintext,omitempty"`
	Prefix          string              `mapstructure:"prefix,omitempty"` // title prefix
	ChangeMode      ChangeDetectionMode `mapstructure:"change_mode"`
	MessageTemplate string              `mapstructure:"msg_template"`
}

func (t *ChangeDetectionMode) UnmarshalMapstructure(input any) error {
	str, ok := input.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", input)
	}

	if str == "" {
		*t = DefaultChangeDetectionMode
		return nil
	}

	mappedVal, ok := changeModeMap[strings.TrimSpace(str)]

	if !ok {
		return fmt.Errorf("Invalid value for change detection mode: %s", str)
	}

	*t = mappedVal
	return nil
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
