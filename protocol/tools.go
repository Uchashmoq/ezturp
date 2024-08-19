package protocol

import (
	"encoding/binary"
	"fmt"
)

func uint32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, n)
	return b
}

func bytesToInt32(p []byte) uint32 {
	u := binary.BigEndian.Uint32(p)
	return u
}

func intToBytes(n int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(n))
	return b
}

func bytesToInt(p []byte) int {
	u := binary.BigEndian.Uint32(p)
	return int(u)
}

func int64ToBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(n))
	return b
}

func bytesToInt64(p []byte) int64 {
	u := binary.BigEndian.Uint64(p)
	return int64(u)
}

func BytesFormat(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float32(bytes)/1024.0)
	} else {
		return fmt.Sprintf("%.2f MB", float32(bytes)/1024.0/1024.0)
	}
}
