package workermanGoClient

import (
	"fmt"
	"testing"
	"time"
)

func TestGetAllGatewayAddressFromRegister(t *testing.T) {
	client := NewInstance("127.0.0.1", 1238, time.Second*30, "")
	fmt.Println(client)
	clientIdList := client.GetAllClientSessions("")
	fmt.Println(clientIdList)
}

func TestGetClientIdByUid(t *testing.T) {
	client := NewInstance("127.0.0.1", 1238, time.Second*30, "")
	clientList := client.GetClientIdByUid("1")
	fmt.Println(clientList)
}

func TestSendToUid(t *testing.T) {
	client := NewInstance("127.0.0.1", 1238, time.Second*30, "")

	for i := 0; i < 1000; i++ {
		client.SendToUid("1", fmt.Sprintf("Hello %d", i))
	}
}
