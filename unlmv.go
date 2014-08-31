package main

import (
	"bytes"
	"encoding/json"
	common "github.com/citruspi/Ludicrous-MV-Common"
	"github.com/hinasssan/msgpack-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	app     = kingpin.New("unlmv", "Hash based compression client.")
	targets = kingpin.Arg("target", "File to encode").Required().Strings()
	log     = logrus.New()
)

const (
	v               = true
	REGISTER string = "http://localhost:5688"
)

func Decode(input string, token bool) {

	lmv_file := new(common.LMVFile)

	if token {

		download_address := REGISTER + "/files/" + input + "/"

		resp, err := http.Get(download_address)

		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(body, &lmv_file)

		if err != nil {
			log.Fatal(err)
		}

		if v {
			log.WithFields(logrus.Fields{
				"token": input,
			}).Info("Retrieved data using token")
		}

	} else {

		file, err := os.Open(input)

		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()

		stat, err := file.Stat()

		if err != nil {
			log.Fatal(err)
		}

		bs := make([]byte, stat.Size())
		_, err = file.Read(bs)

		if err != nil {
			log.Fatal(err)
		}

		err = msgpack.Unmarshal(bs, lmv_file)

		if err != nil {
			log.Fatal(err)
		}

		if v {
			log.WithFields(logrus.Fields{
				"file": input,
			}).Info("Unpacked .lmv file")
		}

	}

	bs := bytes.NewBuffer(make([]byte, 0))

	for i := 0; i < len(lmv_file.Chunks); i++ {

		chunk := make([]byte, lmv_file.Chunks[i].Size)

		f, err := os.Open("/dev/urandom")

		for {

			_, err = f.Read(chunk)

			if err != nil {
				log.Fatal(err)
			}

			if bytes.Equal([]byte(lmv_file.Chunks[i].Hash), []byte(common.CalculateSHA512(chunk))) {
				break
			}

		}

		bs.Write(chunk)

		if v {
			log.WithFields(logrus.Fields{
				"chunk#": i + 1,
			}).Info("Rebuilt chunk")
		}

	}

	fo, err := os.Create(lmv_file.Name)

	if err != nil {
		log.Fatal(err)
	}

	if _, err := fo.Write(bs.Bytes()); err != nil {
		log.Fatal(err)
	}

	if v {
		log.WithFields(logrus.Fields{
			"file": lmv_file.Name,
		}).Info("Writing output to file")
	}

}

func main() {

	kingpin.Parse()

	for _, target := range *targets {
		if _, err := os.Stat(target); err == nil {
			Decode(target, false)
		} else {
			Decode(target, true)
		}
	}

}
func init() {
	log.Formatter = new(logrus.TextFormatter)
}
