package main

import "net/url"

type Api interface {
	createApiUrl(params url.Values) (apiUrl string)
	Get(container interface{}, params url.Values)
}

type ApiCreate interface {
	Api
	Create(body interface{}, container interface{})
}
