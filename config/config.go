package config

import (
	"encoding/json"
	"io/ioutil"
)

var Conf Config

func LoadConfig() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &Conf); err != nil {
		panic(err)
	}
}
