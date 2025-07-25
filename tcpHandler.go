package proxy

import (
	"crypto/tls"
	"errors"
	"github.com/xyjwsj/request-proxy/model"
	"log"
	"net"
)

// HandleTCP 处理 TCP/WebSocket 连接
func HandleTCP(wrapReq model.WrapRequest) {
	defer wrapReq.Conn.Close()

	// 1. 读取目标地址（例如 CONNECT 请求中的 Host 或 TLS ClientHello 中的 SNI）
	targetAddr, err := net.ResolveTCPAddr("tcp", "0")
	if err != nil {
		log.Println("解析tcp代理目标地址错误：" + err.Error())
		return
	}

	// 2. 连接到目标服务器
	serverConn, err := net.DialTCP("tcp", nil, targetAddr)
	if err != nil {
		log.Printf("连接到 %s 失败: %v\n", targetAddr, err)
		return
	}
	defer serverConn.Close()

	log.Printf("建立隧道: 客户端 <-> 代理 <-> %s", targetAddr)

	// 2. 根据 SNI 获取或生成子证书
	host, port, _ := net.SplitHostPort(wrapReq.Conn.RemoteAddr().String())
	cert, err := Cache.GetCertificate(host, port)
	if err != nil {
		log.Printf("Failed to get certificate for %s: %v", host, err)
		_, _ = wrapReq.Writer.WriteString("HTTP/1.1 502 Bad Gateway\r\n\r\n")
		_ = wrapReq.Writer.Flush()
		return
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS10,
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			host := hello.ServerName
			cert, err := Cache.GetCertificate(host, "443")
			if err != nil {
				return nil, err
			}
			if certInfo, ok := cert.(tls.Certificate); ok {
				return &certInfo, nil
			}
			return nil, errors.New("invalid certificate type")
		},
		Certificates: []tls.Certificate{cert.(tls.Certificate)},
	}

	// 3. 开始 TLS 握手
	sslConn := tls.Server(wrapReq.Conn, tlsConfig)
	err = sslConn.Handshake()
	if err != nil {
		log.Printf("TLS handshake failed: %v", err)
		return
	}

	// 3. 双向转发数据
	errChan := make(chan error, 2)

	//go transferData(wrapReq.Conn, serverConn, errChan)
	//go transferData(serverConn, wrapReq.Conn, errChan)

	// 等待任意一方关闭
	<-errChan
}
