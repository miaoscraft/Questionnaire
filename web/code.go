package web

import (
	"github.com/satori/go.uuid"
	"sync"
)

//Codes 储存已验证的验证码
var Codes struct {
	sync.Mutex
	M map[string]Player
	U map[Player]string
}

//Player 代表一个玩家的昵称和UUID
type Player struct {
	Name string
	ID   uuid.UUID
}

func init() {
	Codes.M = make(map[string]Player)
	Codes.U = make(map[Player]string)
}
