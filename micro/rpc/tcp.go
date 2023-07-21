package rpc

import (
	"encoding/binary"
	"net"
)

var (
	numOfLengthBytes = 8
)

func ReadMsg(conn net.Conn) ([]byte, error) {
	lenBs := make([]byte, numOfLengthBytes)
	_, err := conn.Read(lenBs)
	if err != nil {
		return nil, err
	}
	headerLength := binary.BigEndian.Uint32(lenBs[:4])
	bodyLength := binary.BigEndian.Uint32(lenBs[4:])
	reqBs := make([]byte, headerLength+bodyLength)
	_, err = conn.Read(reqBs[8:])
	if err != nil {
		return nil, err
	}
	copy(reqBs[:8], lenBs)
	return reqBs, nil
}
