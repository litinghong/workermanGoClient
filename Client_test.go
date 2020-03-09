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
			client.SendToUid("1", fmt.Sprintf("7SO3lsyD8D68INzetZCQqjy6L8B4xPQCf68AM4PlN5gyjEjE0Sc9VJGI44SCVRoOduppEzksRIRyCyyfaHDkrpxa76USkHrJVZ6T0uJGIrNyB6uIJcwjH4jS8M8oSaEvFk9FiSGTgAQP2ax8TL8nsuzgA5Y0VX3mfIBofJyIh5PB0B4jfQwXMpWJXN34GkjQkm000YONVziINf1JwXQJIifEDWerSd2hX65TvJ4b7mzqrwptN1uOThwDLsyL2ONnguqIQyrZubvJjzdNBXJFRjudopGDmhnM1ZChoOuQwWjHg4LGRWnnPiVw1rX8GpXyEbJ6bOPovdsdm2exjCAx2p7m7gpUDb9roJ0LnlKE9lJGQJU8fKv5B5ALmxMvcYegSk0CslMQ24Vg52PtkvtEP4TgOMDetvys1V3zg2mPuN9ZoOfFZqCVKwMFaho7prOFCT8wbpqRrj0P5mLoIw4HPOxR7ofRhWLEmIgFNdPAaMi9nXBFOEHw9o3dVwli7UV%d", i))
		}
		wg.Done()
	}()

	go func() {
		for j := 100001; j <= 200000; j++ {
			client.SendToUid("2", fmt.Sprintf("7SO3lsyD8D68INzetZCQqjy6L8B4xPQCf68AM4PlN5gyjEjE0Sc9VJGI44SCVRoOduppEzksRIRyCyyfaHDkrpxa76USkHrJVZ6T0uJGIrNyB6uIJcwjH4jS8M8oSaEvFk9FiSGTgAQP2ax8TL8nsuzgA5Y0VX3mfIBofJyIh5PB0B4jfQwXMpWJXN34GkjQkm000YONVziINf1JwXQJIifEDWerSd2hX65TvJ4b7mzqrwptN1uOThwDLsyL2ONnguqIQyrZubvJjzdNBXJFRjudopGDmhnM1ZChoOuQwWjHg4LGRWnnPiVw1rX8GpXyEbJ6bOPovdsdm2exjCAx2p7m7gpUDb9roJ0LnlKE9lJGQJU8fKv5B5ALmxMvcYegSk0CslMQ24Vg52PtkvtEP4TgOMDetvys1V3zg2mPuN9ZoOfFZqCVKwMFaho7prOFCT8wbpqRrj0P5mLoIw4HPOxR7ofRhWLEmIgFNdPAaMi9nXBFOEHw9o3dVwli7UV%d", j))
		}
		wg.Done()
	}()

	go func() {
		for k := 200001; k <= 300000; k++ {
			client.SendToUid("3", fmt.Sprintf("7SO3lsyD8D68INzetZCQqjy6L8B4xPQCf68AM4PlN5gyjEjE0Sc9VJGI44SCVRoOduppEzksRIRyCyyfaHDkrpxa76USkHrJVZ6T0uJGIrNyB6uIJcwjH4jS8M8oSaEvFk9FiSGTgAQP2ax8TL8nsuzgA5Y0VX3mfIBofJyIh5PB0B4jfQwXMpWJXN34GkjQkm000YONVziINf1JwXQJIifEDWerSd2hX65TvJ4b7mzqrwptN1uOThwDLsyL2ONnguqIQyrZubvJjzdNBXJFRjudopGDmhnM1ZChoOuQwWjHg4LGRWnnPiVw1rX8GpXyEbJ6bOPovdsdm2exjCAx2p7m7gpUDb9roJ0LnlKE9lJGQJU8fKv5B5ALmxMvcYegSk0CslMQ24Vg52PtkvtEP4TgOMDetvys1V3zg2mPuN9ZoOfFZqCVKwMFaho7prOFCT8wbpqRrj0P5mLoIw4HPOxR7ofRhWLEmIgFNdPAaMi9nXBFOEHw9o3dVwli7UV%d", k))
		}
		wg.Done()
	}()

	wg.Wait()
	client.Close()
}

func TestSendToAll(t *testing.T) {
	client := NewInstance("127.0.0.1", 1238, time.Second*30, "")
	client.SendToAll("sendToAll", nil, nil, false)
	client.SendToAll("sendToAllWithClientId", []string{"7f0000010b5500000001", "7f0000010b5600000001"}, nil, false)
}

func TestClientToAddress(t *testing.T) {
	localIp, localPort, connId, err := clientIdToAddress("7f0000010b5500000001")
	fmt.Printf("localIp:%s, localPort:%d, connId:%d, err:%s", localIp, localPort, connId, err)
}

func TestGetBufferFromGateway(t *testing.T) {
	protocol := Protocol{}
	protocol.Cmd = CMD_SELECT
	protocol.ExtData = `{"fields":["uid","groups","session"],"where":{"uid":["1"]}}`
	buff, _ := protocol.ToBuffer()
	gatewayBuffer, err := getBufferFromGateway("127.0.0.1:2903", buff)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(gatewayBuffer.UnMarshal())
}
