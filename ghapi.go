package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	// "net/http/httputil"
	"net/url"
)

type GHApi struct {
	UserName string
	Token    string
	APIBase
}

func (api GHApi) NewRequest(method string, body io.Reader, container interface{}) {
	client := &http.Client{}
	url := api.createApiUrl(nil)
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", "token "+api.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	log.Println(method, url)

	// dump, err := httputil.DumpRequestOut(req, true)

	// log.Printf("YEEE %s\n\n", dump)
	// log.Println(err)

	res, err := client.Do(req)

	handleHTTPCall(res, err, &container)
}

func (api GHApi) Get(container interface{}, params url.Values) {
	api.NewRequest("GET", nil, container)
}

func (api GHApi) Create(body interface{}, container interface{}) {
	text, err := json.Marshal(body)

	if err != nil {
		log.Fatal(err)
	}

	api.NewRequest("POST", bytes.NewReader(text), container)
}

func (api GHApi) Edit(body interface{}, container interface{}) {
	text, err := json.Marshal(body)

	if err != nil {
		log.Fatal(err)
	}

	api.NewRequest("PATCH", bytes.NewReader(text), container)
}

type GHAPI interface {
	Api
	Create(body interface{}, container interface{})
	Edit(body interface{}, container interface{})
}

func CreateGHApi(endPoint string) GHApi {
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
