package proxy

import (
	"log"
	"net"
	"request-proxy/util"
)

// HandleClient 处理客户端连接
func HandleClient(clientConn net.Conn) {
	defer clientConn.Close()

	// 解析请求的第一行
	buffer := make([]byte, 1024)
	n, err := clientConn.Read(buffer)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	// 判断是 HTTP/HTTPS 请求还是 WebSocket/TCP 连接
	if util.IsHTTPRequest(buffer[:n]) {
		// HTTP(S) 请求
		HandleHTTP(clientConn, buffer[:n])
	} else {
		// TCP/WebSocket 连接
		HandleTCP(clientConn, buffer[:n])
	}
}
