package main

import (
	"net/url"
)

func CreateSFApi(project string, endPoint string) Api {
	return APIBase{
		Root:     "https://sourceforge.net/rest/p/" + project,
		EndPoint: endPoint,
	}
}

func CallSFAPI(project string, endPoint string, params url.Values, container interface{}) {
	api := CreateSFApi(project, endPoint)

	api.Get(container, params)
}
