package workermanGoClient

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/techoner/gophp/serialize"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type Instance struct {
	secretKey       string
	gatewaySockets  map[string]string
	registerAddress string
	registerPort    int
	connectTimeout  time.Duration
}

type GatewayBuffer struct {
	GatewayAddress string
	Buffer         []byte
}

// 在线用户的session clientId
type ClientWithSession struct {
	GatewayAddress string
	ClientId       uint32
	Session        string
}

type Client struct {
	uid           string
	localIp       string
	localPort     string
	connectionIds []uint32
}

// 网关地址
var gatewayAddress []string

func NewInstance(registerAddr string, port int, connectTimeOut time.Duration, secretKey string) *Instance {
	client := &Instance{
		registerAddress: registerAddr,
		registerPort:    port,
		connectTimeout:  connectTimeOut,
		secretKey:       secretKey,
	}

	err := client.getAllGatewayAddressFromRegister()
	if err != nil {
		fmt.Println(err)
	}
	return client
}

/**
 * 向所有客户端连接(或者 client_id_array 指定的客户端连接)广播消息
 *
 * @param string $message           向客户端发送的消息
 * @param array  $client_id_array   客户端 id 数组
 * @param array  $exclude_client_id 不给这些client_id发
 * @param bool   $raw               是否发送原始数据（即不调用gateway的协议的encode方法）
 * @return void
 * @throws Exception
 */
func (c *Instance) SendToAll(message string, clientIdList []string, raw bool) {
	if len(clientIdList) == 0 {
		return
	}

	gatewayData := &Protocol{}
	gatewayData.Cmd = CMD_SEND_TO_ALL
	gatewayData.Body = message

	if raw == true {
		gatewayData.Flag |= FLAG_NOT_CALL_ENCODE
	}

	dataArray := make(map[string][]uint32)
	for _, clientId := range clientIdList {
		localIp, port, connectionId, err := clientIdToAddress(clientId)
		if err != nil {
			fmt.Println(err)
			continue
		}

		address := fmt.Sprintf("%s:%d", localIp, port)
		dataArray[address] = append(dataArray[address], connectionId)

		//err := sendToGateway(address, gatewayData, c.connectTimeout)
		//if err != nil {
		//	fmt.Println("send to address: ", address, " error: ", err)
		//}
	}

	var success, failure int
	for address, connectionIdList := range dataArray {
		extData := map[string][]uint32{
			"connections": connectionIdList,
		}

		extDataStr, err := json.Marshal(extData)
		if err != nil {
			fmt.Println(err)
			continue
		}
		gatewayData.ExtData = string(extDataStr)
		err = sendToGateway(address, gatewayData, c.connectTimeout)
		if err != nil {
			failure++
			fmt.Println(err)
		} else {
			success++
		}
	}

	fmt.Println("success:", success, "failure:", failure)
}

// 通讯地址到clientId的转换
func clientIdToAddress(clientId string) (localIp string, localPort uint16, connectionId uint32, err error) {
	if len(clientId) != 20 {
		err = errors.New("client_id GatewayBuffer is invalid")
		return
	}

	br := bytes.NewReader([]byte(clientId))

	ipByte := make([]byte, 4)
	binary.Read(br, binary.BigEndian, ipByte)
	ipv4 := net.IPv4(ipByte[0], ipByte[1], ipByte[2], ipByte[3])
	localIp = ipv4.String()

	binary.Read(br, binary.BigEndian, localPort)
	binary.Read(br, binary.BigEndian, connectionId)
	return
}

// 发送数据到网关
func sendToGateway(address string, gatewayData *Protocol, timeout time.Duration) error {
	buffer, err := gatewayData.ToBuffer()
	if err != nil {
		return err
	}
	return sendBufferToGateway(address, buffer, timeout)
}

// 发送buffer数据到网关
func sendBufferToGateway(address string, gatewayBuffer []byte, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("sendBufferToGateway: connected!")
	defer conn.Close()

	// 发送连接信息
	_, err = conn.Write(gatewayBuffer)
	if err != nil {
		fmt.Println("sendBufferToGateway error:", err)
	}

	return nil
}

// 发送buffer数据到网关
func sendAndRecv(address string, gatewayBuffer []byte, recv chan GatewayBuffer) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println(err)
		close(recv)
		return
	}
	fmt.Println(" sendBufferToGateway: ", address, " connected!")
	defer conn.Close()

	// 准备接收信息
	// go handleConn(conn, recv)

	// 发送信息
	_, err = conn.Write(gatewayBuffer)
	if err != nil {
		fmt.Println("sendBufferToGateway error:", err)
	}

	// 接收信息
	var packLen uint32 = 655350
	var realLen uint32
	var packRead bool
	for {
		// read from the connection
		var buf = make([]byte, packLen)
		log.Println("start to read from conn")
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("conn read error:", err)
			close(recv)
			return
		}
		log.Printf("read %d bytes, content is %s\n", n, string(buf[:n]))

		if packRead == false {
			realLen = binary.BigEndian.Uint32(buf[0:4])
		}

		if realLen < packLen {
			// 返回数据
			recv <- GatewayBuffer{address, buf[:n]}
			return
		}
	}
}

func (c *Instance) getAllGatewayAddressFromRegister() error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", c.registerAddress, c.registerPort), c.connectTimeout)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("connected!")
	defer conn.Close()

	ch := make(chan []string)
	go handleRecvAddress(conn, ch)

	// 发送连接信息
	content := fmt.Sprintf("{\"event\":\"worker_connect\",\"secret_key\":\"%s\"}\n", c.secretKey)
	_, err = conn.Write([]byte(content))
	if err != nil {
		fmt.Println("发送连接信息失败:", err)
	}
	select {
	case ack := <-ch: // 接收到服务器返回数据
		gatewayAddress = ack
		fmt.Println("接收到服务器返回数据")
		return nil
	case <-time.After(c.connectTimeout): // 超时
		return errors.New(fmt.Sprint("read time out:", c.connectTimeout))
	}
}

// 接收网关地址
func handleRecvAddress(conn net.Conn, ch chan []string) {
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

		type addresses struct {
			Event     string
			Addresses []string `json:"addresses"`
		}

		var addressEv addresses
		err = json.Unmarshal(buf[:n], &addressEv)
		if err != nil {
			log.Println("服务返回数据解析失败", err)
		}

		if addressEv.Event == "broadcast_addresses" {
			if err != nil {
				log.Println("服务返回数据解析失败", err)
				return
			}
			ch <- addressEv.Addresses
			return
		}
	}
}

func handleConn(conn net.Conn, ch chan []byte) {
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

		// 反回数据
		ch <- buf[:n]
		return
	}
}

// 获取所有在线client_id的session和clientId
func (c *Instance) GetAllClientSessions(group string) []ClientWithSession {
	gatewayData := &Protocol{}
	if group == "" {
		gatewayData.Cmd = CMD_GET_ALL_CLIENT_INFO
	} else {
		gatewayData.Cmd = CMD_GET_CLINET_INFO_BY_GROUP
		gatewayData.ExtData = group
	}

	recv := make(chan GatewayBuffer)
	buffer, _ := gatewayData.ToBuffer()
	for _, address := range gatewayAddress {
		go sendAndRecv(address, buffer, recv)
	}

	var clientIdList []ClientWithSession
	for i := 0; i < len(gatewayAddress); i++ {
		gatewayBuffer := <-recv
		address := gatewayBuffer.GatewayAddress
		response := gatewayBuffer.Buffer

		cs := ClientWithSession{
			GatewayAddress: address,
			ClientId:       0,
		}

		clientIds, err := serialize.UnMarshal(response[4:])
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch clientIds.(type) {
		case map[string]interface{}:
			for i, j := range clientIds.(map[string]interface{}) {
				id, _ := strconv.Atoi(i)
				cs.ClientId = uint32(id)
				if j != nil {
					cs.Session = j.(string)
				}

				clientIdList = append(clientIdList, cs)
			}
		}
	}

	return clientIdList
}

func (c *Instance) GetClientIdByUid(uid string) (clientList []Client) {
	gatewayData := Protocol{}
	gatewayData.Cmd = CMD_GET_CLIENT_ID_BY_UID
	gatewayData.ExtData = uid

	recv := make(chan GatewayBuffer)
	buffer, _ := gatewayData.ToBuffer()
	for _, address := range gatewayAddress {
		go sendAndRecv(address, buffer, recv)
	}

	for i := 0; i < len(gatewayAddress); i++ {
		gatewayBuffer := <-recv
		fmt.Println(string(gatewayBuffer.Buffer))
		addr := strings.Split(gatewayBuffer.GatewayAddress, ":")
		response, err := serialize.UnMarshal(gatewayBuffer.Buffer[4:])
		if err != nil {
			fmt.Println(err)
			continue
		}

		if response == nil {
			continue
		}

		client := Client{
			uid:       uid,
			localIp:   addr[0],
			localPort: addr[1],
		}

		switch response.(type) {
		case []interface{}:
			for _, i := range response.([]interface{}) {
				if i != nil {
					client.connectionIds = append(client.connectionIds, uint32(i.(int)))
				}
			}
		}

		if len(client.connectionIds) > 0 {
			clientList = append(clientList, client)
		}
	}

	return
}

// 向指定uid发送消息
func (c *Instance) SendToUid(uid, message string) {
	gatewayData := Protocol{}
	gatewayData.Cmd = CMD_SEND_TO_UID
	gatewayData.ExtData = fmt.Sprintf("[%s]", uid)
	gatewayData.Body = message

	for _, address := range gatewayAddress {
		err := sendToGateway(address, &gatewayData, time.Second*30)
		if err != nil {
			fmt.Println(err)
		}
	}
}
