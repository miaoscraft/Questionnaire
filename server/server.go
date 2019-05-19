package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	pk "github.com/Tnze/gomcbot/packet"
	"github.com/miaoscraft/SignUpServer/chat"
	"github.com/satori/go.uuid"
)

//OnPlayer 在正版玩家连接时被调用，其返回值将被作为断开连接的原因被发送给客户端
//返回值应当为一个JSON Chat值，例如`"msg"`
var OnPlayer func(name string, UUID uuid.UUID, protocol int32) string

const (
	//Threshold 指定了数据传输时最小压缩包大小
	Threshold = 256

	ServerID   = "joinus.miaoscraft.cn"
	ServerPort = 25565
)

//Client 封装了与客户端之间的底层的网络交互
type Client struct {
	conn   net.Conn
	reader io.ByteReader
	io.Writer

	threshold       int
	ProtocolVersion int32
	Name            string
	ID              uuid.UUID
	skin            string
}

//Handle 接受客户端连接
func Handle(conn net.Conn) {
	defer conn.Close()

	c := Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		Writer: conn,
	}

	nextState, err := c.handshake()
	if err != nil {
		// log.Println(err)
		return
	}

	const (
		CheckState  = 1
		PlayerLogin = 2
	)
	switch nextState {
	case CheckState:
		c.status()
	case PlayerLogin:
		log.Println(c.conn.RemoteAddr(), "协议号", c.ProtocolVersion)

		err = c.login()
		if err != nil {
			msg := chat.NewTranslateMsg("multiplayer.disconnect." + err.Error())
			jmsg, err := json.Marshal(msg)
			if err != nil {
				return
			}
			packet := pk.Packet{ID: 0x00, Data: pk.PackString(string(jmsg))}
			c.Write(packet.Pack(0)) //Ignore error
			return
		}

		resp := `"error"`
		if OnPlayer != nil {
			resp = OnPlayer(c.Name, c.ID, c.ProtocolVersion)
		}

		packet := pk.Packet{
			ID:   disconnectID(c.ProtocolVersion),
			Data: pk.PackString(resp),
		}
		c.Write(packet.Pack(c.threshold)) //Ignore error
		return
	}
}

func (c *Client) handshake() (nextState int32, err error) {
	p, err := pk.RecvPacket(c.reader, false)
	if err != nil {
		return -1, err
	}
	if p.ID != 0 {
		return -1, fmt.Errorf("packet ID 0x%X is not handshake", p.ID)
	}

	r := bytes.NewReader(p.Data)
	c.ProtocolVersion, err = pk.UnpackVarInt(r)
	if err != nil {
		return -1, fmt.Errorf("parse protol version fail: %v", err)
	}

	var (
		sid string
		spt int16
	)
	// ServerID
	sid, err = pk.UnpackString(r)
	if err != nil {
		return -1, fmt.Errorf("parse address fail: %v", err)
	}
	// Server Port
	spt, err = pk.UnpackInt16(r)
	if err != nil {
		return -1, fmt.Errorf("parse port fail: %v", err)
	}

	//检查服务器ID和端口是否匹配
	if sid != ServerID || uint16(spt) != ServerPort {
		return -1, fmt.Errorf("server address rejected")
	}

	nextState, err = pk.UnpackVarInt(r)
	if err != nil {
		return -1, fmt.Errorf("parse next state fail: %v", err)
	}

	return nextState, nil
}

func (c *Client) status() {
	for i := 0; i < 2; i++ {
		p, err := pk.RecvPacket(c.reader, false)
		if err != nil {
			break
		}

		switch p.ID {
		case 0x00:
			respPack := getStatus()
			c.Write(respPack.Pack(0))
		case 0x01:
			c.Write(p.Pack(0))
		}
	}
}

func getStatus() pk.Packet {
	return pk.Packet{
		ID: 0x00,
		Data: pk.PackString(`
		{
			"version": {
				"name": "1.14.1",
				"protocol": 480
			},
			"players": {
				"max": 1,
				"online": 0,
				"sample": []
			},	
			"description": {
				"text": "喵喵公馆验证码服务器"
			}
		}
		`),
	}
}

func disconnectID(protocal int32) byte {
	switch protocal {
	case 404:
		return 0x1B
	case 477, 480:
		return 0x1A
	default:
		return 0x1A
	}
}
