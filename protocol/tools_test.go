package protocol

import (
	"log"
	"math/rand"
	"os"
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

func Test_logger(t *testing.T) {
	logger := log.New(os.Stdout, "udp_server ", log.Ldate|log.Ltime)
	logger.Printf("nihao")
}
func Test_del(t *testing.T) {
	mp := make(map[int]int)
	mp[1] = 2
	delete(mp, 6)
	delete(mp, 4)
}
