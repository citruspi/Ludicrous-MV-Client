package common

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"log"
	"os"
	"os/user"
	"strconv"

	"github.com/tsuru/config"
)

type File struct {
	Size      int64   `msgpack:"size"`
	Name      string  `msgpack:"name"`
	Algorithm string  `msgpack:"algorithm"`
	Chunks    []Chunk `msgpack:"chunks"`
	Tar       bool    `msgpack:"tar"`
}

type Chunk struct {
	Hash  string `msgpack:"hash"`
	Size  int64  `msgpack:"size"`
	Index int    `msgpack:"index"`
}

type ClientConfiguration struct {
	Chunks struct {
		Size int64
	}
	Tracker struct {
		Address string
	}
}

func ProcessClientConfiguration() ClientConfiguration {

	foundConf := true
	conf := ClientConfiguration{}

	if _, err := os.Stat("lmv.yml"); err == nil {
		config.ReadConfigFile("lmv.yml")
	} else {
		usr, err := user.Current()

		if err != nil {
			log.Fatal(err)
		}

		if _, err := os.Stat(usr.HomeDir + "/lmv.conf"); err == nil {
			config.ReadConfigFile(usr.HomeDir + "/lmv.conf")
		} else {
			if _, err := os.Stat("/etc/lmv.conf"); err == nil {
				config.ReadConfigFile("/etc/lmv.conf")
			} else {
				foundConf = false
			}
		}
	}

	if foundConf {
		chunk_size, _ := config.GetString("chunks:size")

		chunk_size_int, err := strconv.ParseInt(chunk_size, 10, 64)

		if err != nil {
			log.Fatal(err)
		}

		conf.Chunks.Size = chunk_size_int

		tracker_address, _ := config.GetString("tracker:address")
		conf.Tracker.Address = tracker_address
	} else {
		conf.Tracker.Address = "http://localhost:8080"
		conf.Chunks.Size = 1048576
	}

	return conf

}

func CalculateSHA512(data []byte) string {

	var hasher hash.Hash = sha512.New()

	hasher.Reset()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))

}
