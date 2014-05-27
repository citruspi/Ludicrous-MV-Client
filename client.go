package main

import (
	"archive/tar"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/hinasssan/msgpack-go"
)

// CONSTANTS

const CHUNK_SIZE int64 = 1048576

var REGISTER string = ""
var log = logrus.New()
var v bool = false

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

func TarballDirectory(fp string) string {

	f, err := ioutil.TempFile("", "")

	if err != nil {
		log.Fatal(err)
	}

	type WalkedFile struct {
		Path string
		Info os.FileInfo
	}

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	tw := tar.NewWriter(f)

	files := make([]WalkedFile, 0)

	tarball := func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			//
		} else {

			files = append(files, WalkedFile{
				Path: path,
				Info: info,
			})

		}
		return nil

	}

	err = filepath.Walk(fp, tarball)
	if err != nil {
		log.Fatal(err)
	}

	for _, fr := range files {

		hdr := &tar.Header{
			Name: fr.Path,
			Size: fr.Info.Size(),
		}

		err := tw.WriteHeader(hdr)

		if err != nil {
			log.Fatal(err)
		}

		file, err := os.Open(fr.Path)

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

		if _, err := tw.Write([]byte(bs)); err != nil {
			log.Fatal(err)
		}

	}

	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	return f.Name()

}

func encode(fp string, token bool) {

	lmv_file := new(LMVFile)

	lmv_file.Algorithm = "SHA512"

	lmv_file.Name = filepath.Base(fp)

	stat, err := os.Stat(fp)

	if err != nil {
		log.Fatal(err)
	}

	if stat.IsDir() {
		fp = TarballDirectory(fp)
		lmv_file.Tar = true
		if v {
			log.WithFields(logrus.Fields{
				"directory": lmv_file.Name,
			}).Info("Beginning new encode")
		}
	} else {
		lmv_file.Tar = false
		if v {
			log.WithFields(logrus.Fields{
				"file": lmv_file.Name,
			}).Info("Beginning new encode")
		}
	}

	file, err := os.Open(fp)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	stat, err = file.Stat()

	if err != nil {
		log.Fatal(err)
	}

	bs := make([]byte, stat.Size())
	_, err = file.Read(bs)

	if err != nil {
		log.Fatal(err)
	}

	lmv_file.Size = stat.Size()

	chunks := make([]LMVChunk, 1)

	if stat.Size() <= CHUNK_SIZE {

		chunks[0] = LMVChunk{
			CalculateSHA512(bs),
			stat.Size(),
			0,
		}

		if v {
			log.WithFields(logrus.Fields{
				"count": 1,
			}).Info("Breaking into chunks")
		}

	} else {

		chunk_count := stat.Size()/CHUNK_SIZE + 1

		if v {
			log.WithFields(logrus.Fields{
				"count": chunk_count - 1,
			}).Info("Breaking into chunks")
		}

		chunks = make([]LMVChunk, chunk_count)

		for i := 0; i < len(chunks)-1; i++ {

			chunk := bs[int64(i)*CHUNK_SIZE : int64(i+1)*CHUNK_SIZE]

			if v {
				log.WithFields(logrus.Fields{
					"chunk#": i + 1,
				}).Info("Encoding chunk")
			}

			chunks[i] = LMVChunk{
				CalculateSHA512(chunk),
				CHUNK_SIZE,
				i,
			}

		}

		chunk := bs[int64(cap(chunks)-1)*CHUNK_SIZE:]

		chunks[cap(chunks)-1] = LMVChunk{
			CalculateSHA512(chunk),
			int64(len(chunk)),
			cap(chunks) - 1,
		}

	}

	lmv_file.Chunks = chunks

	if token {

		upload_address := REGISTER + "/upload"

		packed, err := json.Marshal(lmv_file)

		if err != nil {
			log.Fatal(err)
		}

		fields := make(url.Values)
		fields.Set("file", string(packed))

		resp, err := http.PostForm(upload_address, fields)

		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		if v {
			log.WithFields(logrus.Fields{
				"token": string(body),
			}).Info("Retrieved token from server")
		}

		fmt.Println("'" + lmv_file.Name + "'" + " --> " + "'" + string(body) + "'")

	} else {

		os.Create(lmv_file.Name + ".lmv")

		b, err := msgpack.Marshal(lmv_file)

		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(lmv_file.Name+".lmv", b, 0644)

		if err != nil {
			log.Fatal(err)
		}

		if v {
			log.WithFields(logrus.Fields{
				"file": lmv_file.Name + ".lmv",
			}).Info("Writing output to file")
		}

	}

}

func decode(input string, token bool) {

	lmv_file := new(LMVFile)

	if token {

		download_address := REGISTER + "/download/" + input

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

	start := time.Now()

	token := flag.Bool("token", false, "Use tokens in place of .lmv files")
	verbose := flag.Bool("verbose", false, "Provide verbose output")
	register := flag.String("register", "http://127.0.0.1:8081", "Register for tokens (including protocol)")

	REGISTER = *register

	flag.Parse()

	if *verbose {
		v = true
	} else {
		v = false
	}

	if v {

		log.WithFields(logrus.Fields{
			"token":    *token,
			"verbose":  *verbose,
			"register": *register,
		}).Info("Parsed flags.")

	}

	if len(os.Args) < 2 {

		fmt.Println("Use lmv -h for usage")

	} else {

		paths := strings.Split(os.Args[0], "/")
		exec := paths[len(paths)-1]

		if exec == "lmv" {

			for i := 0; i < len(os.Args[1:]); i++ {

				if _, err := os.Stat(os.Args[i+1]); err == nil {

					encode(os.Args[i+1], *token)

				}

			}

		} else if exec == "unlmv" {

			for i := 0; i < len(os.Args[1:]); i++ {

				if os.Args[i+1] != "--verbose" {

					if _, err := os.Stat(os.Args[i+1]); err == nil {

						decode(os.Args[i+1], false)

					} else {

						decode(os.Args[i+1], true)

					}

				}

			}

		}

	}

	if v {
		elapsed := time.Since(start)

		log.WithFields(logrus.Fields{
			"time": string(elapsed),
		}).Info("Completed execution.")
	}

}