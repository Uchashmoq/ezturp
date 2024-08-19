package protocol

import (
	"log"
	"math/rand"
	"testing"
)

func Test_intToBytes(t *testing.T) {
	id := rand.Int()
	log.Printf("id %v", id)
	idb := intToBytes(id)
	log.Printf("idb %v", idb)
	id1 := bytesToInt(idb)
	log.Printf("id1:%v", id1)
}
