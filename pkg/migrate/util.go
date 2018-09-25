package migrate

import (
	"github.com/ghodss/yaml"
	"github.com/platform9/cctl/pkg/state"
	"log"
)

func EncodeMigratedState(any interface{}) []byte {
	buf, err := yaml.Marshal(any)
	if err != nil {
		log.Fatal("encode:", err)
	}
	return buf
}


func DecodeMigratedState(any []byte) state.State {
	var thisState state.State
	err := yaml.Unmarshal(any, &thisState)
	if err != nil {
		log.Fatal("decode:", err)
	}
	return thisState
}