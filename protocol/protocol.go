package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var (
	HEAD = []byte("yxh")
)

const (
	NEW_SESSION = iota
	REMOVE_SESSION
	DATA
	KEEP_ALIVE
	MAINTAIN_UDP_CLIENT_ADDR
)

/*
HEAD TYPE SESSION DATA_LEN DATA
*/

func Readn(reader io.Reader, n int) ([]byte, error) {
	buf := make([]byte, n)
	for k := 0; k < n; {
		d, err := reader.Read(buf[k:])
		if err != nil {
			return nil, err
		}
		k += d
	}
	return buf, nil
}

func WriteFrame(writer io.Writer, t byte, id uint32, data []byte) (err error) {
	frame := EncodeFrame(t, id, data)
	_, err = writer.Write(frame)
	return err
}

func ParseFrame(p []byte) (t byte, id uint32, data []byte, err error) {
	var offset int
	if len(p) < len(HEAD)+1+4+4 {
		err = errors.New("bad frame")
		return
	}
	headBuf := p[offset : offset+len(HEAD)]
	if !bytes.Equal(headBuf, HEAD) {
		return 0, 0, nil, fmt.Errorf("unknown protocol head '%v'", string(headBuf))
	}
	offset += len(HEAD)

	t = p[offset]
	offset++

	idBuf := p[offset : offset+4]
	offset += 4
	id = bytesToInt32(idBuf)

	lenBuf := p[offset : offset+4]
	offset += 4
	dataLen := bytesToInt(lenBuf)

	if len(p) < len(HEAD)+1+4+4+dataLen {
		err = errors.New("bad frame")
		return
	}

	data = p[offset : offset+dataLen]
	return

}

func ReadFrame(reader io.Reader) (t byte, id uint32, data []byte, err error) {
	headBuf, err := Readn(reader, len(HEAD))
	if err != nil {
		return 0, 0, nil, err
	}
	if !bytes.Equal(headBuf, HEAD) {
		return 0, 0, nil, fmt.Errorf("unknown protocol head '%v'", string(headBuf))
	}

	tBuf, err := Readn(reader, 1)
	if err != nil {
		return 0, 0, nil, err
	}

	t = tBuf[0]

	idBuf, err := Readn(reader, 4)
	if err != nil {
		return 0, 0, nil, err
	}
	id = bytesToInt32(idBuf)
	dataLen, err := Readn(reader, 4)
	if err != nil {
		return 0, 0, nil, err
	}
	data, err = Readn(reader, bytesToInt(dataLen))
	return t, id, data, err
}

func EncodeFrame(t byte, id uint32, data []byte) []byte {
	buf := bytes.Buffer{}
	buf.Write(HEAD)
	buf.Write([]byte{t})
	buf.Write(uint32ToBytes(id))
	buf.Write(uint32ToBytes(uint32(len(data))))
	buf.Write(data)
	return buf.Bytes()
}
