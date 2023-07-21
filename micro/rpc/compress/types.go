package compress

type Compress interface {
	Code() uint8
	Compress(val []byte) ([]byte, error)
	UnCompress(data []byte) ([]byte, error)
}

type NoCompress struct {
}

func (n *NoCompress) Code() uint8 {
	return 1
}

func (n *NoCompress) Compress(data []byte) ([]byte, error) {
	return data, nil
}

func (n *NoCompress) UnCompress(data []byte) ([]byte, error) {
	return data, nil
}
