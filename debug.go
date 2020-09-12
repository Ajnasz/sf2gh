package main

import "fmt"

func debug(args ...interface{}) {
	if cliConfig.debug {
		fmt.Println(args...)
	}
}

func verbose(args ...interface{}) {
	if cliConfig.verbose {
		fmt.Println(args...)
	}
}
