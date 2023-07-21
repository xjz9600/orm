package message

import (
	"bytes"
	"encoding/binary"
)

type Request struct {
	HeadLength uint32
	BodyLength uint32
	RequestID  uint32
	Version    uint8
	Compressor uint8
	Serializer uint8

	ServiceName string
	MethodName  string
	Meta        map[string]string

	Data []byte
}

func EncodeReq(req *Request) []byte {
	bs := make([]byte, req.HeadLength+req.BodyLength)
	binary.BigEndian.PutUint32(bs[:4], req.HeadLength)
	binary.BigEndian.PutUint32(bs[4:8], req.BodyLength)
	binary.BigEndian.PutUint32(bs[8:12], req.RequestID)
	bs[12] = req.Version
	bs[13] = req.Compressor
	bs[14] = req.Serializer
	cur := bs[15:]
	copy(cur, req.ServiceName)
	cur = cur[len(req.ServiceName):]
	cur[0] = '\n'
	cur = cur[1:]
	copy(cur, req.MethodName)
	cur = cur[len(req.MethodName):]
	cur[0] = '\n'
	cur = cur[1:]
	for k, v := range req.Meta {
		copy(cur, k)
		cur = cur[len(k):]
		cur[0] = '\r'
		cur = cur[1:]
		copy(cur, v)
		cur = cur[len(v):]
		cur[0] = '\n'
		cur = cur[1:]
	}
	copy(cur, req.Data)
	return bs

}

func DecodeReq(data []byte) *Request {
	req := &Request{}
	req.HeadLength = binary.BigEndian.Uint32(data[:4])
	req.BodyLength = binary.BigEndian.Uint32(data[4:8])
	req.RequestID = binary.BigEndian.Uint32(data[8:12])
	req.Version = data[12]
	req.Compressor = data[13]
	req.Serializer = data[14]
	header := data[15:req.HeadLength]
	index := bytes.IndexByte(header, '\n')
	req.ServiceName = string(header[:index])
	header = header[index+1:]
	index = bytes.IndexByte(header, '\n')
	req.MethodName = string(header[:index])
	header = header[index+1:]

	index = bytes.IndexByte(header, '\n')
	if index != -1 {
		meta := make(map[string]string)
		for index != -1 {
			pair := header[:index]
			pairIndex := bytes.IndexByte(header, '\r')
			key := string(pair[:pairIndex])
			value := string(pair[pairIndex+1 : index])
			meta[key] = value
			header = header[index+1:]
			index = bytes.IndexByte(header, '\n')
		}
		req.Meta = meta
	}
	req.Data = data[req.HeadLength:]
	return req
}

func (req *Request) CalculateHeaderLength() {
	// 不要忘了分隔符
	headLength := 15 + len(req.ServiceName) + 1 + len(req.MethodName) + 1
	for key, value := range req.Meta {
		headLength += len(key)
		// key 和 value 之间的分隔符
		headLength++
		headLength += len(value)
		headLength++
		// 和下一个 key value 的分隔符
	}
	req.HeadLength = uint32(headLength)
}

func (req *Request) CalculateBodyLength() {
	req.BodyLength = uint32(len(req.Data))
}
