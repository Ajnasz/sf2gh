package main

import "fmt"

func debug(args ...interface{}) {
	if cliConfig.debug {
		fmt.Println(args...)
	}
}
