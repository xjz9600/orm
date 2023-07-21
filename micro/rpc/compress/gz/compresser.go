package gz

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

type GzipCompress struct {
}

func (g *GzipCompress) Code() uint8 {
	return 2
}

func (g *GzipCompress) Compress(val []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(val); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (g *GzipCompress) UnCompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}
