package migrate

import (
	"bytes"
	"encoding/gob"
	"log"
)

func EncodeMigratedState(key interface{}) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		log.Fatal("encode:", err)
	}
	return buf.Bytes()
}