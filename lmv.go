package main

import	(
	"os"
	"fmt"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"path/filepath" 
	"flag"
	"net/http" 
	"net/url"
	"io/ioutil" 	
	"log"
	"strconv"
)

type LMVFile struct {
	hash string
	size int64
	name string
}

func CalculateSHA512(str string) string {

	var hasher hash.Hash = sha512.New()

	hasher.Reset()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))

}

func encode(fp string, token bool) {

	lmv_file := new(LMVFile)

	file, err := os.Open(fp)

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

    lmv_file.hash = CalculateSHA512(string(bs))
    lmv_file.size = stat.Size()
    lmv_file.name = filepath.Base(fp)

	if token {

    	upload_address := "http://127.0.0.1:8081/upload"
		
		fields := make(url.Values)
        fields.Set("name", lmv_file.name)
        fields.Set("hash", lmv_file.hash)
        fields.Set("size", strconv.FormatInt(lmv_file.size, 10))
    	
    	resp, err := http.PostForm(upload_address, fields)

    	if err != nil {
			log.Fatal(err)
		}
		
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("'" + lmv_file.name + "'" + " --> " + "'" + string(body) + "'")
	
	} else {

		os.Create(lmv_file.name + ".lmv")

		full := lmv_file.hash + "\n" + lmv_file.name + "\n" + strconv.FormatInt(lmv_file.size, 10)

		err = ioutil.WriteFile(lmv_file.name + ".lmv", []byte(full), 0644)
    	
		if err != nil { 
			log.Fatal(err)
		}

    }

}

func main() {

	token := flag.Bool("token", false, "Use tokens in place of .lmv files")

	flag.Parse()

	if len(os.Args) < 2 {

		fmt.Println("Use lmv -h for usage")

	} else {

		for i := 0; i < len(os.Args[1:]); i++ {
	     
			if _, err := os.Stat(os.Args[i+1]); err == nil {

	    		encode(os.Args[i+1], *token)

			}

	    }

	}
}