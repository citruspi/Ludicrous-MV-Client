package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/hinasssan/msgpack-go"
)

// CONSTANTS

const (
	CHUNK_SIZE int64  = 1048576
	REGISTER   string = "http://localhost:5688"
	v          bool   = true
)

var log = logrus.New()

type LMVFile struct {
	Size      int64      `msgpack:"size"`
	Name      string     `msgpack:"name"`
	Algorithm string     `msgpack:"algorithm"`
	Chunks    []LMVChunk `msgpack:"chunks"`
	Tar       bool       `msgpack:"tar"`
}

type LMVChunk struct {
	Hash  string `msgpack:"hash"`
	Size  int64  `msgpack:"size"`
	Index int    `msgpack:"index"`
}

func CalculateSHA512(data []byte) string {

	var hasher hash.Hash = sha512.New()

	hasher.Reset()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))

}

func decode(input string, token bool) {

	lmv_file := new(LMVFile)

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

			if bytes.Equal([]byte(lmv_file.Chunks[i].Hash), []byte(CalculateSHA512(chunk))) {
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

func init() {
	log.Formatter = new(logrus.TextFormatter)
}

func main() {

	if len(os.Args) < 2 {

		fmt.Println("Use lmv -h for usage")

	} else {

		for i := 0; i < len(os.Args[1:]); i++ {

			if _, err := os.Stat(os.Args[i+1]); err == nil {

				decode(os.Args[i+1], false)

			} else {

				decode(os.Args[i+1], true)

			}

		}

	}

}
