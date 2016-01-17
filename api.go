package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

const apiRoot = "https://sourceforge.net/rest/p/fluxbox"

func createUrl(endPoint string) (newUrl string) {
	baseUrl, err := url.Parse(apiRoot)

	if err != nil {
		log.Fatal(err)
	}

	baseUrl.Path = path.Join(baseUrl.Path, endPoint)

	return baseUrl.String()
}

func CallAPI(endPoint string, container interface{}) {
	res, err := http.Get(createUrl(endPoint))

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
