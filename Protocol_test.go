package workermanGoClient

import (
	"fmt"
	"io/ioutil"
	"net"
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
	buf, _ := p.ToBuffer()
	err := ioutil.WriteFile("/Users/yons/Documents/workermanGoClient/w.bin", buf, os.ModeAppend)
	fmt.Println(err)
}

func TestGatewayDirectConn(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:2900")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for i := 0; i < 1000; i++ {

		gatewayData := Protocol{}
		gatewayData.Cmd = CMD_SEND_TO_UID
		gatewayData.ExtData = fmt.Sprintf(`["%s"]`, "1")
		gatewayData.Body = fmt.Sprintf("7SO3lsyD8D68INzetZCQqjy6L8B4xPQCf68AM4PlN5gyjEjE0Sc9VJGI44SCVRoOduppEzksRIRyCyyfaHDkrpxa76USkHrJVZ6T0uJGIrNyB6uIJcwjH4jS8M8oSaEvFk9FiSGTgAQP2ax8TL8nsuzgA5Y0VX3mfIBofJyIh5PB0B4jfQwXMpWJXN34GkjQkm000YONVziINf1JwXQJIifEDWerSd2hX65TvJ4b7mzqrwptN1uOThwDLsyL2ONnguqIQyrZubvJjzdNBXJFRjudopGDmhnM1ZChoOuQwWjHg4LGRWnnPiVw1rX8GpXyEbJ6bOPovdsdm2exjCAx2p7m7gpUDb9roJ0LnlKE9lJGQJU8fKv5B5ALmxMvcYegSk0CslMQ24Vg52PtkvtEP4TgOMDetvys1V3zg2mPuN9ZoOfFZqCVKwMFaho7prOFCT8wbpqRrj0P5mLoIw4HPOxR7ofRhWLEmIgFNdPAaMi9nXBFOEHw9o3dVwli7UV%d", i)
		buff, err := gatewayData.ToBuffer()

		if err != nil {
			t.Fatal(err)
		}

		n, err := conn.Write(buff)
		if err != nil {
			t.Fatal(err)
		}

		if n != len(buff) {
			t.Fatal("发送长度不一致")
		}
	}
}
