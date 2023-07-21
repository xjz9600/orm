package net

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	go func() {
		err := Serve("tcp", ":8082")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	err := Connect("tcp", "localhost:8082")
	t.Log(err)
}

func TestSend(t *testing.T) {
	server := &Server{}
	go func() {
		err := server.Start("tcp", ":8083")
		t.Log(err)
	}()
	time.Sleep(time.Second * 3)
	client := &Client{
		network: "tcp",
		addr:    "localhost:8083",
	}
	resp, err := client.Send("Hello")
	assert.NoError(t, err)
	assert.Equal(t, resp, "HelloHello")
}
