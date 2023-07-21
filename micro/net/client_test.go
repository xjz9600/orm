package net

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConnect(t *testing.T) {
	type args struct {
		network string
		addr    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, Connect(tt.args.network, tt.args.addr), fmt.Sprintf("Connect(%v, %v)", tt.args.network, tt.args.addr))
		})
	}
}
