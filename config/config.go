package config

import (
	"errors"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Values of the configuration
type Values struct {
	Stripe struct {
		PublicKey     string `yaml:"public_key"`
		SecretKey     string `yaml:"secret_key"`
		WebhookSecret string `yaml:"webhook_secret"`
	} `yaml:"stripe"`
}

// Read the configuration values and rerturn them
func Read() (Values, error) {
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		return Values{}, errors.New("could not read configuration file")
	}
	v := Values{}

	err = yaml.Unmarshal([]byte(data), &v)
	if err != nil {
		return Values{}, errors.New("could not parse configuration file")
	}

	return v, nil
}
