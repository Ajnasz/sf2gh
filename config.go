package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type GithubConfig struct {
	UserName    string `json:"userName"`
	AccessToken string `json:"accessToken"`
}

type Config struct {
	Github GithubConfig
}

func GetConfig() Config {
	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		log.Fatal(err)
	}

	var config Config

	json.Unmarshal(file, &config)

	return config
}
