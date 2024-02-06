package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"tailscale-route-tiller/slack"

	"gopkg.in/yaml.v2"
)

// Config is a struct for our YAML data
type Config struct {
	Subnets           []string `yaml:"subnets"`
	Sites             []string `yaml:"sites"`
	TailscaleCommand  string   `yaml:"TailscaleCommand"`
	EnableIpv6        bool     `yaml:"EnableIpv6"`
	TailscaleclientId string   `yaml:"TailscaleclientId"`
	TailscaleKey      string   `yaml:"TailscaleKey"`
	Slack             Slack    `yaml:"Slack"`
	SQS               SQS      `yaml:"SQS"`
}

type Slack struct {
	WebhookURL string `yaml:"WebhookURL"`
	Enabled    bool   `yaml:"Enabled"`
}

type SQS struct {
	QueueURL string `yaml:"QueueURL"`
	Region   string `yaml:"Region"`
}

var ActiveConfig *Config

// ReadYAML reads the YAML configuration file
func ReadYAML(filename string) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading YAML file: ", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		os.Exit(1)
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		fmt.Println("Error reading YAML file: ", err)
		if ActiveConfig.Slack.Enabled {
			slack.WebhookURL = ActiveConfig.Slack.WebhookURL
			slack.PostError(err)
		}
		os.Exit(1)
	}
	ActiveConfig = c
}
