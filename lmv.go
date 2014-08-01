package main

import (
	"./core"

	"os"

	"gopkg.in/alecthomas/kingpin.v1"
)

var (
	app     = kingpin.New("lmv", "Hash based compression client.")
	token   = kingpin.Flag("token", "Use tokens in place of .lmv files").Bool()
	targets = kingpin.Arg("target", "File to encode").Required().Strings()
)

func main() {

	kingpin.Parse()

	for _, target := range *targets {
		if _, err := os.Stat(target); err == nil {
			core.Encode(target, *token)
		}
	}
}
