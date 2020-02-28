package workermanGoClient

import (
	"testing"
	"time"
)

func TestGetAllGatewayAddressFromRegister(t *testing.T) {
	client := NewClient("127.0.0.1", 1236, time.Second*30, "")
	client.getAllGatewayAddressFromRegister()
}
