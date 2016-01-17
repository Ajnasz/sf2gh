package main

type Api interface {
	createApiUrl() (apiUrl string)
	Get(container interface{})
}
