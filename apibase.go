package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

func handleHTTPCall(res *http.Response, err error, container *interface{}) {

	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal(err)
	}

	// handle if status code is not 200

	err = json.Unmarshal(body, container)

	if res.StatusCode/100 != 2 {
		log.Fatal(string(body), res.Header)
	}

	if err != nil {
		log.Fatal(err)
	}
}

type APIBase struct {
	Root     string
	EndPoint string
}

func (api APIBase) createApiUrl(params url.Values) (apiUrl string) {
	baseUrl, err := url.Parse(api.Root)

	if err != nil {
		log.Fatal(err)
	}

	q := baseUrl.Query()

	for key, _ := range params {
		q.Set(key, params.Get(key))
	}

	baseUrl.RawQuery = q.Encode()

	baseUrl.Path = path.Join(baseUrl.Path, api.EndPoint)

	return baseUrl.String()
}

func (api APIBase) Get(container interface{}, params url.Values) {
	reqUrl := api.createApiUrl(params)
	log.Println("GET", reqUrl)
	res, err := http.Get(reqUrl)

	if err != nil {
		log.Fatal(err)
	}

	handleHTTPCall(res, err, &container)
}

// func (api Api) Create() {
// }

// func (api Api) Update() {
// }

// func (api Api) Delete() {
// }
