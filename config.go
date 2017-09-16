package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// GithubConfig defines github configuration
type GithubConfig struct {
	UserName    string `json:"userName"`
	AccessToken string `json:"accessToken"`
}

// Config configuration struct
type Config struct {
	Github GithubConfig
}

// GetConfig reads config.json
func GetConfig() Config {
	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		log.Fatal(err)
	}

	var config Config

	json.Unmarshal(file, &config)

	return config
}
