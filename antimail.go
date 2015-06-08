package main

import (
	"log"
	"net"

	"github.com/gchaincl/antimail/smtp"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		defer conn.Close()

		s := smtp.NewSession(conn, conn)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Error: %s\n", r.(error).Error())
				}
			}()

			_, err := s.Run()
			if err != nil {
				panic(err)
			}

		}()
		log.Printf("+ %s\n", conn.RemoteAddr().String())
	}
}