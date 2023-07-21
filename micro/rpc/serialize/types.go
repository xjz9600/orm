package serialize

type Serialize interface {
	Code() uint8
	Encode(val any) ([]byte, error)
	Decode(data []byte, val any) error
}
