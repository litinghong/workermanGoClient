package workermanGoClient

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/techoner/gophp/serialize"
)

type Protocol struct {
	PacketLen    uint32
	Cmd          uint8
	LocalIp      uint32
	LocalPort    uint16
	ClientIp     uint32
	ClientPort   uint16
	ConnectionId uint32
	Flag         uint8
	GatewayPort  uint16
	ExtLen       uint32
	ExtData      string
	Body         interface{}
	MixedBody    string
}

const (
	// 包头长度
	HEAD_LEN uint8 = 28
	// 发给worker，gateway有一个新的连接
	CMD_ON_CONNECTION uint8 = 1
	// 发给worker的，客户端有消息
	CMD_ON_MESSAGE uint8 = 3
	// 发给worker上的关闭链接事件
	CMD_ON_CLOSE uint8 = 4
	// 发给gateway的向单个用户发送数据
	CMD_SEND_TO_ONE uint8 = 5
	// 发给gateway的向所有用户发送数据
	CMD_SEND_TO_ALL uint8 = 6
	// 发给gateway的踢出用户
	CMD_KICK uint8 = 7
	// 发给gateway，通知用户session更新
	CMD_UPDATE_SESSION uint8 = 9
	// 获取在线状态
	CMD_GET_ALL_CLIENT_INFO uint8 = 10
	// 判断是否在线
	CMD_IS_ONLINE uint8 = 11
	// client_id绑定到uid
	CMD_BIND_UID uint8 = 12
	// 解绑
	CMD_UNBIND_UID uint8 = 13
	// 向uid发送数据
	CMD_SEND_TO_UID uint8 = 14
	// 根据uid获取绑定的clientid
	CMD_GET_CLIENT_ID_BY_UID uint8 = 15
	// 加入组
	CMD_JOIN_GROUP uint8 = 20
	// 离开组
	CMD_LEAVE_GROUP uint8 = 21
	// 向组成员发消息
	CMD_SEND_TO_GROUP uint8 = 22
	// 获取组成员
	CMD_GET_CLINET_INFO_BY_GROUP uint8 = 23
	// 获取组成员数
	CMD_GET_CLIENT_COUNT_BY_GROUP uint8 = 24
	// worker连接gateway事件
	CMD_WORKER_CONNECT uint8 = 200
	// 心跳
	CMD_PING uint8 = 201
	// GatewayClient连接gateway事件
	CMD_GATEWAY_CLIENT_CONNECT uint8 = 202
	// 根据client_id获取session
	CMD_GET_SESSION_BY_CLIENT_ID uint8 = 203
	// 获取当前绑定的用户列表
	CMD_GET_UIDS uint8 = 205
	// 发给gateway，覆盖session
	CMD_SET_SESSION uint8 = 204
	// 包体是标量
	FLAG_BODY_IS_SCALAR uint8 = 0x01
	// 通知gateway在send时不调用协议encode方法，在广播组播时提升性能
	FLAG_NOT_CALL_ENCODE uint8 = 0x02
)

func NewProtocol(buffer []byte) *Protocol {
	protocol := &Protocol{}

	if len(buffer) > 0 {
		bReader := bytes.NewReader(buffer)
		binary.Read(bReader, binary.BigEndian, &protocol.PacketLen)
		binary.Read(bReader, binary.BigEndian, &protocol.Cmd)
		binary.Read(bReader, binary.BigEndian, &protocol.LocalIp)
		binary.Read(bReader, binary.BigEndian, &protocol.LocalPort)
		binary.Read(bReader, binary.BigEndian, &protocol.ClientIp)
		binary.Read(bReader, binary.BigEndian, &protocol.ClientPort)
		binary.Read(bReader, binary.BigEndian, &protocol.ConnectionId)
		binary.Read(bReader, binary.BigEndian, &protocol.Flag)
		binary.Read(bReader, binary.BigEndian, &protocol.GatewayPort)
		binary.Read(bReader, binary.BigEndian, &protocol.ExtLen)

		if protocol.ExtLen > 0 {
			extData := make([]byte, protocol.ExtLen)
			binary.Read(bReader, binary.BigEndian, &extData)
			protocol.ExtData = string(extData)
		}

		bodyLen := protocol.PacketLen - protocol.ExtLen - 28
		if bodyLen > 0 {
			bodyData := make([]byte, bodyLen)
			binary.Read(bReader, binary.BigEndian, &bodyData)
			protocol.Body = string(bodyData)
		}
	}

	return protocol
}

func (p *Protocol) ToBuffer() ([]byte, error) {
	buffer := bytes.NewBuffer([]byte{})
	p.ExtLen = uint32(len(p.ExtData))

	var body []byte
	switch p.Body.(type) {
	case string:
		body = []byte(fmt.Sprintf("%s", p.Body))
		p.Flag |= 1
		break
	case int:
		body = []byte(fmt.Sprintf("%s", p.Body))
		p.Flag |= 1
		break
	case bool:
		body = []byte(fmt.Sprintf("%s", p.Body))
		p.Flag |= 1
		break
	case float32:
		body = []byte(fmt.Sprintf("%s", p.Body))
		p.Flag |= 1
		break
	case float64:
		body = []byte(fmt.Sprintf("%s", p.Body))
		p.Flag |= 1
		break
	default:
		_body, err := serialize.Marshal(p.Body)
		if err != nil {
			fmt.Println("toBuffer Error:", err)
			return nil, err
		}
		body = _body
	}

	p.PacketLen = 28 + p.ExtLen + uint32(len(body))

	binary.Write(buffer, binary.BigEndian, p.PacketLen)
	binary.Write(buffer, binary.BigEndian, p.Cmd)
	binary.Write(buffer, binary.BigEndian, p.LocalIp)
	binary.Write(buffer, binary.BigEndian, p.LocalPort)
	binary.Write(buffer, binary.BigEndian, p.ClientIp)
	binary.Write(buffer, binary.BigEndian, p.ClientPort)
	binary.Write(buffer, binary.BigEndian, p.ConnectionId)
	binary.Write(buffer, binary.BigEndian, p.Flag)
	binary.Write(buffer, binary.BigEndian, p.GatewayPort)
	binary.Write(buffer, binary.BigEndian, p.ExtLen)
	binary.Write(buffer, binary.BigEndian, []byte(p.ExtData))
	binary.Write(buffer, binary.BigEndian, body)
	return buffer.Bytes(), nil
}
