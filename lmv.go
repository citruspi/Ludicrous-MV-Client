package main

import  (
    "os"
    "fmt"
    "crypto/sha512"
    "encoding/hex"
    "github.com/hinasssan/msgpack-go"
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
    Hash string `msgpack:"hash"`
    Size int64  `msgpack:"size"`
    Name string `msgpack:"name"`
    Algorithm string `msgpack:"algorithm"`
}

func CalculateSHA512(data []byte) string {

    var hasher hash.Hash = sha512.New()

    hasher.Reset()
    hasher.Write(data)
    return hex.EncodeToString(hasher.Sum(nil))

}

func encode(fp string, token bool) {

    lmv_file := new(LMVFile)

    lmv_file.Algorithm = "SHA512"

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

    lmv_file.Hash = CalculateSHA512(bs)
    lmv_file.Size = stat.Size()
    lmv_file.Name = filepath.Base(fp)

    if token {

        upload_address := "http://127.0.0.1:8081/upload"
        
        fields := make(url.Values)
        fields.Set("name", lmv_file.Name)
        fields.Set("hash", lmv_file.Hash)
        fields.Set("size", strconv.FormatInt(lmv_file.Size, 10))
        
        resp, err := http.PostForm(upload_address, fields)

        if err != nil {
            log.Fatal(err)
        }
        
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)

        if err != nil {
            log.Fatal(err)
        }

        fmt.Println("'" + lmv_file.Name + "'" + " --> " + "'" + string(body) + "'")
    
    } else {

        os.Create(lmv_file.Name + ".lmv")

        b, err := msgpack.Marshal(lmv_file)

        if err != nil {
            log.Fatal(err)
        }    

        err = ioutil.WriteFile(lmv_file.Name + ".lmv", b, 0644)
        
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
