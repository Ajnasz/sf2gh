package main

import (
	"net/url"
)

func CreateSFApi(endPoint string) Api {
	return APIBase{
		Root:     "https://sourceforge.net/rest/p/fluxbox",
		EndPoint: endPoint,
	}
}

func CallSFAPI(endPoint string, params url.Values, container interface{}) {
	api := CreateSFApi(endPoint)

	api.Get(container, params)
}
