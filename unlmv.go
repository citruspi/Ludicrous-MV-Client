package main

import (
	"./core"

	"os"

	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	app     = kingpin.New("unlmv", "Hash based compression client.")
	targets = kingpin.Arg("target", "File to encode").Required().Strings()
)

func main() {

	kingpin.Parse()

	for _, target := range *targets {
		if _, err := os.Stat(target); err == nil {
			core.Decode(target, false)
		} else {
			core.Decode(target, true)
		}
	}

}
