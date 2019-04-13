package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/miaoscraft/Questionnaire/server"
	"github.com/miaoscraft/Questionnaire/web"
	"github.com/satori/go.uuid"
)

func init() {
	server.OnPlayer = Response
}

const (
	httpAddr = ":1314"
	mcAddr   = ":25565"
)

func main() {
	go web.Listen(httpAddr) //HTTP server
	log.Println("在" + httpAddr + "启动网页服务器")

	l, err := net.Listen("tcp", mcAddr) //MC server
	log.Println("在" + mcAddr + "启动验证码服务器")

	if err != nil {
		log.Fatal(err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Print(err)
		}

		go server.Handle(c)
	}
}

//Response 返回验证码
func Response(name string, id uuid.UUID, pto int32) string {
	code := randCode()
	p := web.Player{name, id}

	web.Codes.Lock()
	defer web.Codes.Unlock()

	if oldcode, ok := web.Codes.U[p]; ok {
		code = oldcode
	} else {
		web.Codes.U[p] = code
		web.Codes.M[code] = p
		go func() { //30分钟后删除验证码
			time.Sleep(time.Minute * 30)
			web.Codes.Lock()
			delete(web.Codes.U, p)
			delete(web.Codes.M, code)
			web.Codes.Unlock()
		}()
	}

	log.Println(name, id, pto, code)
	return fmt.Sprintf(
		`"您的验证码: §b%s
§8(30分钟内有效)"`, code)
}

func randCode() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
