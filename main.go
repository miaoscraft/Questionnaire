package main

import (
	"github.com/miaoscraft/SignUpServer/server"
	"log"
	"net"
)

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
