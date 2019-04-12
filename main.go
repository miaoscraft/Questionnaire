package main

import (
	"fmt"
	"github.com/miaoscraft/Questionnaire/server"
	"github.com/satori/go.uuid"
	"log"
	"net"
)

func init() {
	server.OnPlayer = Response
}

func main() {
	l, err := net.Listen("tcp", ":25565")
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
