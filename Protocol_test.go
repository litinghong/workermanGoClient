package workermanGoClient

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewProtocol(t *testing.T) {
	bytes, _ := ioutil.ReadFile("/Users/yons/Documents/workermanGoClient/bin")
	NewProtocol(bytes)
}

func TestToBuffer(t *testing.T) {
	p := NewProtocol(nil)
	p.Body = "messageString"
	buf := p.ToBuffer()
	err := ioutil.WriteFile("/Users/yons/Documents/workermanGoClient/w.bin", buf, os.ModeAppend)
	fmt.Println(err)
}
