package main

func CreateSFApi(endPoint string) Api {
	return APIBase{
		Root:     "https://sourceforge.net/rest/p/fluxbox",
		EndPoint: endPoint,
	}
}

func CallSFAPI(endPoint string, container interface{}) {
	api := CreateSFApi(endPoint)
	api.Get(container)
}
