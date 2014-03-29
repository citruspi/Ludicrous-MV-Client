package main

import  (
    "os"
    "fmt"
    "crypto/sha512"
    "encoding/hex"
    "encoding/json"
    "github.com/hinasssan/msgpack-go"
    "hash"
    "path/filepath"
    "flag"
    "net/http"
    "net/url"
    "io/ioutil"
    "log"
)

type LMVFile struct {
    Size int64  `msgpack:"size"`
    Name string `msgpack:"name"`
    Algorithm string `msgpack:"algorithm"`
    Chunks []LMVChunk `msgpack:"chunks"`
}

type LMVChunk struct {
    Hash string `msgpack:"hash"`
    Size int64 `msgpack:"size"`
    Index int `msgpack:"index"`
}

// CONSTANTS

const CHUNK_SIZE int64 = 1048576

func CalculateSHA512(data []byte) string {

    var hasher hash.Hash = sha512.New()

    hasher.Reset()
    hasher.Write(data)
    return hex.EncodeToString(hasher.Sum(nil))

}

func encode(fp string, token bool, register string) {

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

    lmv_file.Size = stat.Size()
    lmv_file.Name = filepath.Base(fp)

    chunks := make([]LMVChunk, 1)

    if stat.Size() <= CHUNK_SIZE {

        chunks[0] = LMVChunk {
            CalculateSHA512(bs),
            stat.Size(),
            0,
        }

    } else {

        chunk_count := stat.Size() / CHUNK_SIZE + 1

        chunks = make([]LMVChunk, chunk_count)

        for i := 0; i < len(chunks) - 1; i++ {

            chunk := bs[int64(i)*CHUNK_SIZE:int64(i+1)*CHUNK_SIZE]

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
            cap(chunks)-1,
        }

    }

    lmv_file.Chunks = chunks

    if token {

        upload_address := register + "/upload"

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
    register := flag.String("register", "http://127.0.0.1:8081", "Register for tokens (including protocol)")

    flag.Parse()

    if len(os.Args) < 2 {

        fmt.Println("Use lmv -h for usage")

    } else {

        for i := 0; i < len(os.Args[1:]); i++ {

            if _, err := os.Stat(os.Args[i+1]); err == nil {

                encode(os.Args[i+1], *token, *register)

            }

        }

    }
}
