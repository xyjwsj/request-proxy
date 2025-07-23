package proxy

import (
	"bufio"
	"fmt"
	"github.com/xyjwsj/request-proxy/model"
	"net"
)

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
		Conn:   clientConn,
		Reader: reader,
		Writer: writer,
	}

	switch peekHex {
	case "0x47", "0x43", "0x50", "0x4f", "0x44", "0x48":
		HandleHTTP(request)
		break
	case "0x5":
		// TODO socket
	default:
		// TODO TCP
	}

	////判断是 HTTP/HTTPS 请求还是 WebSocket/TCP 连接
	//if util.IsHTTPRequest(method, hostPortURL.Opaque) {
	//	// HTTP 请求
	//	HandleHTTP(clientConn, request)
	//} else {
	//	// HTTPS/TCP/WebSocket 连接
	//	HandleTCP(clientConn)
	//}
}
