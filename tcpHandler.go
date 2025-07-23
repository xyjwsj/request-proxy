package proxy

import (
	"bytes"
	"fmt"
	"github.com/xyjwsj/request-proxy/model"
	"github.com/xyjwsj/request-proxy/util"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

// HandleTCP 处理 TCP/WebSocket 连接
func HandleTCP(clientConn net.Conn) {
	if clientConn == nil {
		return
	}
	defer clientConn.Close()

	// 尝试判断是否为 WebSocket 握手
	if util.IsWebSocketHandshake(nil) {
		log.Println("Handling WebSocket connection")
		//handleWebSocket(clientConn, data)
		return
	}

	var b [1024]byte
	n, err := clientConn.Read(b[:])
	if err != nil {
		log.Println(err)
		return
	}

	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == "443" { //https访问
		address = hostPortURL.Scheme + ":443"
	} else { //http访问
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	} //获得了请求的host和port，就开始拨号吧

	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	if method == "CONNECT" {
		fmt.Fprint(clientConn, "HTTP/1.1 200 Connection established\r\n\r\n")
		//io.Copy(client, server)
		clientBuf := make([]byte, 65535)
		written, _ := io.CopyBuffer(model.WrapWriter{Writer: server}, clientConn, clientBuf)
		log.Println("客户端发送服务器:" + string(clientBuf[:written]))
	} else {
		server.Write(b[:n])
		buf, _ := io.ReadAll(server)
		clientConn.Write(buf)
	} //进行转发
}
