package proxy

import (
	"github.com/xyjwsj/request-proxy/util"
	"log"
	"net"
	"net/url"
	"testing"
)

func TestRequest(t *testing.T) {
	certificate := util.NewCertificate()
	certificate.Init()
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

func TestUrlParse(t *testing.T) {
	parse, _ := url.Parse("CONNECTplatform.hoolai.com:443 HTTP/1.1")
	log.Println(parse)
}
