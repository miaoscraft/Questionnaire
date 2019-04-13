package main

import (
	"fmt"
	"log"
	"net"

	"github.com/miaoscraft/Questionnaire/server"
	"github.com/miaoscraft/Questionnaire/web"
	"github.com/satori/go.uuid"
)

func init() {
	server.OnPlayer = Response
}

func main() {
	go web.Listen(":1314") //HTTP server

	l, err := net.Listen("tcp", ":25565") //MC server
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
	log.Println(name, id, pto)

	return fmt.Sprintf(`"%s"`, name)
}
