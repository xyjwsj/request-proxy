package proxy

import (
	"bufio"
	"fmt"
	"github.com/xyjwsj/request-proxy/model"
	"github.com/xyjwsj/request-proxy/util"
	"net"
)

type ConfigProxy struct {
	requestCall  model.RequestCall
	responseCall model.ResponseCall
	https        bool
}

var config *ConfigProxy

func init() {
	config = &ConfigProxy{
		requestCall:  nil,
		responseCall: nil,
		https:        false,
	}
}

func ConfigHttps(https bool) {
	config.https = https
}

func ConfigOnRequest(onRequest model.RequestCall) {
	config.requestCall = onRequest
}

func ConfigOnResponse(onResponse model.ResponseCall) {
	config.responseCall = onResponse
}

// HandleClient 处理客户端连接
func HandleClient(clientConn net.Conn) {
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)
	writer := bufio.NewWriter(clientConn)

	peek, err := reader.Peek(1)
	if err != nil {
		return
	}

	peekHex := fmt.Sprintf("0x%x", peek[0])

	request := model.WrapRequest{
		ID:     util.UUID(),
		Conn:   clientConn,
		Reader: reader,
		Writer: writer,
	}

	if config != nil {
		if config.requestCall != nil {
			request.OnRequest = config.requestCall
		}
		if config.responseCall != nil {
			request.OnResponse = config.responseCall
		}
		request.Https = config.https
	}

	switch peekHex {
	case "0x47", "0x43", "0x50", "0x4f", "0x44", "0x48":
		HandleHTTP(request)
		break
	case "0x5":
		// TODO socket
	default:
		// TODO TCP
		HandleTCP(request)
	}
}
