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

const (
	//Threshold 指定了数据传输时最小压缩包大小
	Threshold = 256
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
			log.Println(err)
			msg := chat.NewTranslateMsg("multiplayer.disconnect." + err.Error())
			jmsg, err := json.Marshal(msg)
			if err != nil {
				return
			}
			packet := pk.Packet{ID: 0x00, Data: pk.PackString(string(jmsg))}
			c.Write(packet.Pack(0)) //Ignore error
			return
		}

		packet := pk.Packet{
			ID: 0x1B,
			Data: pk.PackString(`
			{
				"text":"您的验证码:WHDS"
			}
			`),
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

	// ServerID
	_, err = pk.UnpackString(r)
	if err != nil {
		return -1, fmt.Errorf("parse address fail: %v", err)
	}
	// Server Port
	_, err = pk.UnpackInt16(r)
	if err != nil {
		return -1, fmt.Errorf("parse port fail: %v", err)
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
				"name": "1.13.2",
				"protocol": 404
			},
			"players": {
				"max": 1,
				"online": 0,
				"sample": [
					{
						"name": "Tnze",
						"id": "58f6356e-b30c-4811-8bfc-d72a9ee99e73"
					}
				]
			},	
			"description": {
				"text": "喵喵公馆验证码服务器"
			}
		}
		`),
	}
}
