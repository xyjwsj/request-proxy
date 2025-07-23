package proxy

import (
	"log"
	"net"
	"testing"
)

func TestRequest(t *testing.T) {
	ln, _ := net.Listen("tcp", ":8080")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept Error:" + err.Error())
			continue
		}
		go HandleClient(conn)
	}
}
