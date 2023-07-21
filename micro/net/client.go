package net

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func Connect(network, addr string) error {
	conn, err := net.DialTimeout(network, addr, time.Second*3)
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	for i := 0; i < 10; i++ {
		_, err = conn.Write([]byte("Hello"))
		if err != nil {
			return err
		}
		res := make([]byte, 16)
		_, err := conn.Read(res)
		if err != nil {
			return err
		}
		fmt.Println(string(res))
	}
	return nil
}

type Client struct {
	network string
	addr    string
}

func (c *Client) Send(data string) (string, error) {
	conn, err := net.DialTimeout(c.network, c.addr, time.Second*3)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = conn.Close()
	}()

	res := make([]byte, len(data)+numOfLengthBytes)
	binary.BigEndian.PutUint64(res[:8], uint64(len(data)))
	copy(res[8:], data)
	_, err = conn.Write(res)
	if err != nil {
		return "", err
	}
	lenBs := make([]byte, numOfLengthBytes)
	_, err = conn.Read(lenBs)
	if err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint64(lenBs)
	reqBs := make([]byte, length)
	_, err = conn.Read(reqBs)
	if err != nil {
		return "", err
	}
	return string(reqBs), nil
}
