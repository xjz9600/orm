package message

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRespEncodeDecode(t *testing.T) {
	testcase := []struct {
		name string
		resp *Response
	}{
		{
			name: "normal",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compressor: 13,
				Serializer: 14,
				Error:      []byte("this is error"),
				Data:       []byte("hello, world"),
			},
		},
		{
			name: "data with \n",
			resp: &Response{
				RequestID:  123,
				Version:    12,
				Compressor: 13,
				Serializer: 14,
				Data:       []byte("hello \n world"),
			},
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			tc.resp.CalculateHeaderLength()
			tc.resp.CalculateBodyLength()
			data := EncodeResp(tc.resp)
			req := DecodeResp(data)
			assert.Equal(t, req, tc.resp)
		})
	}
}
