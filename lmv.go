package main

import (
	"archive/tar"
	"encoding/json"
	lmv "github.com/citruspi/Ludicrous-MV-Common"
	"github.com/franela/goreq"
	"github.com/hinasssan/msgpack-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"os"
	"path/filepath"
)

var (
	app     = kingpin.New("lmv", "Hash based compression client.")
	token   = kingpin.Flag("token", "Use tokens in place of .lmv files").Bool()
	targets = kingpin.Arg("target", "File to encode").Required().Strings()
	log     = logrus.New()
)

const (
	v                 = true
	CHUNK_SIZE int64  = 1048576
	REGISTER   string = "http://localhost:5688"
)

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

func Encode(fp string, token bool) {

	lmv_file := new(lmv.File)

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

	chunks := make([]lmv.Chunk, 1)

	if stat.Size() <= CHUNK_SIZE {

		chunks[0] = lmv.Chunk{
			lmv.CalculateSHA512(bs),
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

		chunks = make([]lmv.Chunk, chunk_count)

		for i := 0; i < len(chunks)-1; i++ {

			chunk := bs[int64(i)*CHUNK_SIZE : int64(i+1)*CHUNK_SIZE]

			if v {
				log.WithFields(logrus.Fields{
					"chunk#": i + 1,
				}).Info("Encoding chunk")
			}

			chunks[i] = lmv.Chunk{
				lmv.CalculateSHA512(chunk),
				CHUNK_SIZE,
				i,
			}

		}

		chunk := bs[int64(cap(chunks)-1)*CHUNK_SIZE:]

		chunks[cap(chunks)-1] = lmv.Chunk{
			lmv.CalculateSHA512(chunk),
			int64(len(chunk)),
			cap(chunks) - 1,
		}

	}

	lmv_file.Chunks = chunks

	if token {

		res, err := goreq.Request{
			Method:      "POST",
			Uri:         REGISTER + "/files/",
			ContentType: "application/json",
			Body:        lmv_file,
		}.Do()

		if err != nil {
			log.Fatal(err)
		}

		if v {
			log.Info("POST'ed data to the token server")
		}

		parsed := map[string]interface{}{}
		response, err := res.Body.ToString()

		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal([]byte(response), &parsed)

		if err != nil {
			log.Fatal(err)
		}

		if v {
			log.WithFields(logrus.Fields{
				"token": parsed["token"],
			}).Info("Retrieved token from response")
		}

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

func main() {

	kingpin.Parse()

	for _, target := range *targets {
		if _, err := os.Stat(target); err == nil {
			Encode(target, *token)
		}
	}
}

func init() {
	log.Formatter = new(logrus.TextFormatter)
}
