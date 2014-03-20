package main

import	(
	"os"
	"fmt"
	"crypto/sha512"
	"encoding/hex"
	"hash"
	"path/filepath" 
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

func encode(fp string) {

	lmv_file := new(LMVFile)

	file, err := os.Open(fp)

    if err != nil {
    	// handle error
        return
    }

    defer file.Close()	

    stat, err := file.Stat()
    if err != nil {
        return
    }

    bs := make([]byte, stat.Size())
    _, err = file.Read(bs)
    if err != nil {
        return
    }

    str := string(bs)
    //fmt.Println(str)

    lmv_file.hash = CalculateSHA512(str)
    lmv_file.size = stat.Size()
    lmv_file.name = filepath.Base(fp)

    fmt.Println("Hash: ", lmv_file.hash)
 	fmt.Println("Size: ", lmv_file.size)   
 	fmt.Println("Name: ", lmv_file.name)

}


func main() {

	if len(os.Args) < 2 {

		fmt.Println("Usage")

	} else {

		for i := 0; i < len(os.Args[1:]); i++ {
	     
			if _, err := os.Stat(os.Args[i+1]); err == nil {

	    		encode(os.Args[i+1])

			}

	    }

	}
}