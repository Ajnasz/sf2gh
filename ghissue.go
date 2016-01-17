package main

import (
	"io"
	"log"
	"net/http"
)

type GHApi struct {
	UserName string
	Token    string
	APIBase
}

func (api GHApi) NewRequest(method string, body io.Reader, container interface{}) {
	client := &http.Client{}
	url := api.createApiUrl()
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "token "+api.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	log.Println(url)

	res, err := client.Do(req)

	handleHTTPCall(res, err, &container)
}

func (api GHApi) Get(container interface{}) {
	api.NewRequest("GET", nil, container)
}

func CreateGHApi(endPoint string) Api {
	config := GetConfig()

	return GHApi{
		config.Github.UserName,
		config.Github.AccessToken,
		APIBase{
			Root:     "https://api.github.com",
			EndPoint: endPoint,
		},
	}
}
