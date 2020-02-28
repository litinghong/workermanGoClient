package workermanGoClient

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type Client struct {
	secretKey       string
	gatewaySockets  map[string]string
	gatewayAddress  []string
	registerAddress string
	registerPort    int
	connectTimeout  time.Duration
}

type event struct {
	Event string `json:"event"`
}

func NewClient(registerAddr string, port int, connectTimeOut time.Duration, secretKey string) *Client {
	client := &Client{
		registerAddress: registerAddr,
		registerPort:    port,
		connectTimeout:  connectTimeOut,
		secretKey:       secretKey,
	}

	return client
}

func (c *Client) getAllGatewayAddressFromRegister() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.registerAddress, c.registerPort), c.connectTimeout)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("connected!")
	defer conn.Close()

	go c.handleConn(conn)

	// 发送连接信息
	content := fmt.Sprintf("{\"event\":\"worker_connect\",\"secret_key\":\"%s\"}\n", c.secretKey)
	_, err = conn.Write([]byte(content))
	if err != nil {
		fmt.Println("发送连接信息失败:", err)
	}

	select {}
	return nil
}

func (c *Client) handleConn(conn net.Conn) {
	defer conn.Close()
	for {
		// read from the connection
		var buf = make([]byte, 655350)
		log.Println("start to read from conn")
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("conn read error:", err)
			return
		}
		log.Printf("read %d bytes, content is %s\n", n, string(buf[:n]))

		var ev event
		err = json.Unmarshal(buf[:n], &ev)
		if err != nil {
			log.Println("服务返回数据解析失败", err)
		}

		if ev.Event == "broadcast_addresses" {
			type addresses struct {
				Addresses []string `json:"addresses"`
			}
			var addr addresses
			err = json.Unmarshal(buf[:n], &addr)
			if err != nil {
				log.Println("服务返回数据解析失败", err)
				return
			}
			c.gatewayAddress = addr.Addresses
			return
		}
	}
}

func (c *Client) Close() {

}
