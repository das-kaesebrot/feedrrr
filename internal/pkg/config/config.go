package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type YamlURL struct {
	*url.URL
}

type FeedrrrConfig struct {
	Jobs  map[string]JobConfig `yaml:"jobs"`
	Sinks map[string][]string  `yaml:"sinks"`
}

type JobConfig struct {
	Sinks        []string `yaml:"sinks"`
	CronSchedule string   `yaml:"schedule"`
	FeedSource   YamlURL  `yaml:"source"`
}

func (u *YamlURL) UnmarshalYAML(value *yaml.Node) error {
	url, err := url.Parse(value.Value)
	u.URL = url
	if err != nil {
		return err
	}
	if u.URL.Scheme != "" && u.URL.Host != "" {
		return fmt.Errorf("URL is not valid!")
	}
	return nil
}

func (u YamlURL) MarshalYAML() (any, error) {
	return u.String(), nil
}

func ParseConfig(c *FeedrrrConfig, appName string) error {
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
		return err
	}

	err = viper.Unmarshal(c)
	if err != nil {
		return err
	}

	return nil
}
