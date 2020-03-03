package workermanGoClient

import (
	"fmt"
	"sync"
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

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		for i := 1; i <= 100000; i++ {
			client.SendToUid("1", fmt.Sprintf("%d", i))
		}
		wg.Done()
	}()

	go func() {
		for j := 100001; j <= 200000; j++ {
			client.SendToUid("2", fmt.Sprintf("%d", j))
		}
		wg.Done()
	}()

	go func() {
		for k := 200001; k <= 300000; k++ {
			client.SendToUid("3", fmt.Sprintf("%d", k))
		}
		wg.Done()
	}()

	wg.Wait()
	client.Close()
}
