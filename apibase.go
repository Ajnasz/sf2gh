package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

func handleHTTPCall(res *http.Response, err error, container interface{}) {
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(body, container)

	if err != nil {
		log.Fatal(err)
	}
}

type APIBase struct {
	Root     string
	EndPoint string
}

func (api APIBase) createApiUrl() (apiUrl string) {
	baseUrl, err := url.Parse(api.Root)

	if err != nil {
		log.Fatal(err)
	}

	baseUrl.Path = path.Join(baseUrl.Path, api.EndPoint)

	return baseUrl.String()
}

func (api APIBase) Get(container interface{}) {
	res, err := http.Get(api.createApiUrl())

	handleHTTPCall(res, err, &container)
}

// func (api Api) Create() {
// }

// func (api Api) Update() {
// }

// func (api Api) Delete() {
// }
