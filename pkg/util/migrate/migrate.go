package util

import (
	"bytes"
	"encoding/gob"
	"log"
	"github.com/platform9/cctl/pkg/state"
)

func DecodeMigratedState(any []byte) state.State {
	buf := bytes.NewBuffer(any)
	dec := gob.NewDecoder(buf)
	var thisState state.State
	err := dec.Decode(&thisState)
	if err != nil {
		log.Fatal("decode:", err)
	}
	return thisState
}