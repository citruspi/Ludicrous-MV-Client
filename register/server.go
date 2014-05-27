package main

import (
	"crypto/rand"
	"flag"
	"log"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type LMVFile struct {
	Size      int64      `bson:"size"         json:"size"         binding:"required"`
	Name      string     `bson:"name"         json:"name"         binding:"required"`
	Algorithm string     `bson:"algorithm"    json:"algorithm"    binding:"required"`
	Chunks    []LMVChunk `bson:"chunks"       json:"chunks"       binding:"required"`
	Tar       bool       `bson:"tar"          json:"tar"`
	Token     string     `bson:"token"        json:"token"`
}

type LMVChunk struct {
	Hash  string `bson:"hash"         json:"hash"         binding:"required"`
	Size  int64  `bson:"size"         json:"size"         binding:"required"`
	Index int    `bson:"index"        json:"index"        binding:"required"`
}

const (
	token_length = 10
	mgo_host     = "localhost"
	mgo_db       = "Ludicrous-MV"
	mgo_col      = "Files"
)

func randstr(length int) string {

	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var bytes = make([]byte, length)

	rand.Read(bytes)

	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}

	return string(bytes)

}

func main() {

	host := flag.String("host", "127.0.0.1", "")
	port := flag.String("port", "5688", "")

	flag.Parse()

	session, err := mgo.Dial(mgo_host)

	if err != nil {
		panic(err)
	}

	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	c := session.DB(mgo_db).C(mgo_col)

	m := martini.Classic()
	m.Use(render.Renderer())

	m.Get("/file/:token", func(params martini.Params, r render.Render) {

		n, err := c.Find(bson.M{"token": params["token"]}).Count()

		if err != nil {
			log.Fatal(err)
		}

		if n != 1 {
			r.JSON(404, "")
		} else {
			var lmv_file LMVFile

			err = c.Find(bson.M{"token": params["token"]}).One(&lmv_file)

			if err != nil {
				log.Fatal(err)
			}

			r.JSON(200, lmv_file)
		}

	})

	m.Post("/file", binding.Bind(LMVFile{}), func(params martini.Params, r render.Render, lmv_file LMVFile) {

		token := randstr(token_length)

		for {
			n, err := c.Find(bson.M{"token": token}).Count()

			if err != nil {
				log.Fatal(err)
			}

			if n > 0 {
				token = randstr(token_length)
			} else {
				break
			}
		}

		lmv_file.Token = token

		err := c.Insert(lmv_file)

		if err != nil {
			log.Fatal(err)
		}

		r.JSON(200, map[string]interface{}{"token": token})

	})

	log.Fatal(http.ListenAndServe(*host+":"+*port, m))
}
