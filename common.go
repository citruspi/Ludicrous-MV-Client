package common

import (
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

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
