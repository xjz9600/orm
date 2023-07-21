package net

import (
	"encoding/binary"
	"net"
)

var (
	numOfLengthBytes = 8
)

func Serve(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if err := handleConn(conn); err != nil {
				conn.Close()
			}
		}()
	}
}

func handleConn(conn net.Conn) error {
	for {
		bs := make([]byte, 8)
		_, err := conn.Read(bs)
		if err != nil {
			return err
		}
		//if n != 8 {
		//	return errors.New("micro：没有读到数据")
		//}
		res := handleMsg(bs)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
		//if n != 8 {
		//	return errors.New("micro：没有写完数据")
		//}
	}
}

func handleMsg(data []byte) []byte {
	res := make([]byte, 2*len(data))
	copy(res[:len(data)], data)
	copy(res[len(data):], data)
	return res
}

type Server struct {
}

func (s *Server) Start(network, addr string) error {
	listener, err := net.Listen(network, addr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			if err := s.handleConnV1(conn); err != nil {
				conn.Close()
			}
		}()
	}
}

func (s *Server) handleConnV1(conn net.Conn) error {
	for {
		lenBs := make([]byte, numOfLengthBytes)
		_, err := conn.Read(lenBs)
		if err != nil {
			return err
		}
		length := binary.BigEndian.Uint64(lenBs)
		reqBs := make([]byte, length)
		_, err = conn.Read(reqBs)
		if err != nil {
			return err
		}
		respData := handleMsg(reqBs)
		res := make([]byte, len(respData)+numOfLengthBytes)
		binary.BigEndian.PutUint64(res[:8], uint64(len(respData)))
		copy(res[8:], respData)
		_, err = conn.Write(res)
		if err != nil {
			return err
		}
	}
}
