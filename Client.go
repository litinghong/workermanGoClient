package workermanGoClient

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/techoner/gophp/serialize"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
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

type gatewaySocket struct {
	id             int
	GatewayAddress string
	IsIdle         bool
	conn           net.Conn
}

// 网关地址
var gatewayAddress []string

// 连接池
var (
	pool map[string][]*gatewaySocket
	// 连接池写锁
	poolRWGuard sync.Mutex
	// 连接池中当前连接数量
	socketCount int
)

// 连接池最大连接数量
const PoolMaxConn = 100

// 连接超时时间(秒)
const ConnectionTimeout = 30

var (
	socketId      int
	socketIdGuard sync.Mutex
)

func generateSocketId() int {
	socketIdGuard.Lock()
	socketId++
	socketIdGuard.Unlock()
	return socketId
}

// 新建连接
func newSocket(address string) (*gatewaySocket, error) {
	conn, err := net.DialTimeout("tcp", address, time.Second*ConnectionTimeout)
	if err != nil {
		return nil, err
	}

	socket := &gatewaySocket{
		GatewayAddress: address,
		IsIdle:         true,
		conn:           conn,
	}
	socket.id = generateSocketId()
	return socket, nil
}

func getSocket(address string) (*gatewaySocket, error) {
	poolRWGuard.Lock()
	defer poolRWGuard.Unlock()
	for _, socket := range pool[address] {
		if socket.IsIdle == true {
			socket.IsIdle = false
			return socket, nil
		}
	}

	if len(pool[address]) < PoolMaxConn {
		socket, err := newSocket(address)
		if err != nil {
			return nil, err
		}

		fmt.Println("sendBufferToGateway: ", address, " connected!", socketCount)

		if pool == nil {
			pool = make(map[string][]*gatewaySocket)
		}

		pool[address] = append(pool[address], socket)
		socketCount++
		socket.IsIdle = false
		return socket, nil
	}

	return nil, errors.New("已超过最大连接数")
}

// 从连接池中移除socket
func removeSocket(socket *gatewaySocket) {
	poolRWGuard.Lock()
	newPool := make(map[string][]*gatewaySocket)

	for _, s := range pool[socket.GatewayAddress] {
		if s.id != socket.id {
			newPool[socket.GatewayAddress] = append(newPool[socket.GatewayAddress], s)
		}
	}
	pool = newPool
	poolRWGuard.Unlock()
}

// 尝试恢复连接
func recoverSocket(socket *gatewaySocket) error {
	socket.conn.Close()
	conn, err := net.DialTimeout("tcp", socket.GatewayAddress, time.Second*ConnectionTimeout)
	socket.conn = conn
	return err
}

func (s *gatewaySocket) free() {
	s.IsIdle = true
}

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

func (c *Instance) GetAllUidList() {

}

func selectFromGateway(fields []string, where map[string]interface{}) {
	//gatewayData := Protocol{}
	//gatewayData.Cmd = CMD_SELECT
	//extData := map[string]interface{}{
	//	"fields": fields,
	//	"where": where,
	//}
	//
	//if where["client_id"] != nil {
	//	clientIdLIst := where["client_id"]
	//
	//}

}

// 向指定uid发送消息
func (c *Instance) SendToUid(uid, message string) {
	gatewayData := Protocol{}
	gatewayData.Cmd = CMD_SEND_TO_UID
	gatewayData.ExtData = fmt.Sprintf("[%s]", uid)
	gatewayData.Body = message

	for _, address := range gatewayAddress {
		err := sendToGateway(address, &gatewayData)
		if err != nil {
			fmt.Println(err)
		}
	}
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
func (c *Instance) SendToAll(message string, clientIdList, excludeIdList []string, raw bool) {
	gatewayData := &Protocol{}
	gatewayData.Cmd = CMD_SEND_TO_ALL
	gatewayData.Body = message

	if raw == true {
		gatewayData.Flag |= FLAG_NOT_CALL_ENCODE
	}

	excludeDict := map[string]bool{}
	for _, s := range excludeIdList {
		excludeDict[s] = true
	}

	dataArray := make(map[string][]uint32)

	var success, failure int
	if clientIdList != nil {
		for _, clientId := range clientIdList {
			if excludeDict[clientId] == true {
				continue
			}

			localIp, port, connectionId, err := clientIdToAddress(clientId)
			if err != nil {
				fmt.Println(err)
				continue
			}

			address := fmt.Sprintf("%s:%d", localIp, port)
			dataArray[address] = append(dataArray[address], connectionId)
		}

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
			err = sendToGateway(address, gatewayData)
			if err != nil {
				failure++
				fmt.Println(err)
			} else {
				success++
			}
		}
	} else if len(excludeIdList) == 0 {
		sendToAllGateway(gatewayData)
	} else {
		exclude := map[string]map[uint32]uint32{}
		for _, clientId := range excludeIdList {
			localIp, localPort, connectionId, err := clientIdToAddress(clientId)
			if err != nil {
				address := fmt.Sprintf("%s:%d", localIp, localPort)
				exclude[address][connectionId] = connectionId
			}
		}

		for _, gatewayAddress := range gatewayAddress {
			if exclude[gatewayAddress] != nil {
				extData, err := json.Marshal(map[string]map[uint32]uint32{
					"exclude": exclude[gatewayAddress],
				})

				if err == nil {
					fmt.Println(err)
					continue
				}
				gatewayData.ExtData = string(extData)
				err = sendToGateway(gatewayAddress, gatewayData)
				fmt.Println(err)
			}
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

	dest := make([]byte, hex.DecodedLen(20))
	_, err = hex.Decode(dest, []byte(clientId))
	if err != nil {
		return
	}

	localIp = fmt.Sprintf("%d.%d.%d.%d", dest[0], dest[1], dest[2], dest[3])
	localPort = binary.BigEndian.Uint16(dest[4:6])
	connectionId = binary.BigEndian.Uint32(dest[6:])

	return
}

// 发送数据到网关
func sendToGateway(address string, gatewayData *Protocol) error {
	buffer, err := gatewayData.ToBuffer()
	if err != nil {
		return err
	}
	return sendBufferToGateway(address, buffer)
}

// 向所有 gateway 发送数据
func sendToAllGateway(gatewayData *Protocol) []error {
	var errs []error
	for _, address := range gatewayAddress {
		err := sendToGateway(address, gatewayData)
		if err != nil {
			fmt.Println(err)
			errs = append(errs, err)
		}
	}

	return errs
}

// 发送buffer数据到网关
func sendBufferToGateway(address string, gatewayBuffer []byte) error {
	socket, err := getSocket(address)
	if err != nil {
		fmt.Println(err)
		return err
	}
	conn := socket.conn

	// 发送连接信息
	i, err := conn.Write(gatewayBuffer)
	if err != nil {
		fmt.Println("sendBufferToGateway error:", err)

		err = recoverSocket(socket)
		if err != nil {
			fmt.Println("sendBufferToGateway try to recover socket failure")

			// 移除无效连接
			removeSocket(socket)
		} else {
			// 再重试一次
			i, err = socket.conn.Write(gatewayBuffer)
			if err != nil {
				fmt.Println("sendBufferToGateway recover send error:", err)
			}
		}
	}

	if i < len(gatewayBuffer) {
		fmt.Println("sendBufferToGateway 发送长度不对:", len(gatewayBuffer), i)
	}

	socket.free()
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

// 关闭连接池中所有连接
func (c *Instance) Close() {
	for _, sockets := range pool {
		for _, socket := range sockets {
			err := socket.conn.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
