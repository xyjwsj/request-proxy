package proxy

import (
	"log"
	"net"
	"request-proxy/util"
)

// HandleTCP 处理 TCP/WebSocket 连接
func HandleTCP(clientConn net.Conn, data []byte) {
	// 假设目标地址
	targetAddr := "target.example.com:80" // 替换为目标地址

	// 连接到目标服务器
	serverConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer serverConn.Close()

	// 先发送缓存的数据
	_, err = serverConn.Write(data)
	if err != nil {
		log.Println("Write error:", err)
		return
	}

	// 双向转发数据
	done := make(chan struct{})
	go util.CopyData(clientConn, serverConn, done)
	go util.CopyData(serverConn, clientConn, done)
	<-done
	<-done
}
