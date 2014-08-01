package main

import (
	"./core"

	"os"

	"gopkg.in/alecthomas/kingpin.v1"
)

func main() {

	app := kingpin.New("lmv-client", "Hash based compression client.")
	encodem := app.Command("encode", "Encode mode")
	token := encodem.Flag("token", "Use tokens in place of .lmv files").Bool()
	targete := encodem.Arg("target", "File to encode").Required().Strings()
	decodem := app.Command("decode", "Decode mode")
	targetd := decodem.Arg("target", "File/token to decode").Required().Strings()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {

	case "encode":

		for _, target := range *targete {
			if _, err := os.Stat(target); err == nil {
				core.Encode(target, *token)
			}
		}

	case "decode":

		for _, target := range *targetd {
			if _, err := os.Stat(target); err == nil {
				core.Decode(target, false)
			} else {
				core.Decode(target, true)
			}
		}

	}

}
