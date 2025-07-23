package util

import (
	"bytes"
	"github.com/xyjwsj/request-proxy/model"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// IsHTTPRequest 判断是否为 HTTP 请求
func IsHTTPRequest(method, opaque string) bool {
	return opaque == "443" || method == "GET "
}

func ISHttpsRequest(url *url.URL) bool {
	return strings.HasSuffix(url.Host, ":443") || strings.HasPrefix(url.Scheme, "https")
}

func IsWebSocketHandshake(data []byte) bool {
	return bytes.Contains(data, []byte("Upgrade: websocket"))
}

// HandleHTTP 处理 HTTP 请求
func HandleHTTP(clientConn net.Conn, data []byte) {
	// 创建反向代理
	targetURL, err := url.Parse("http://example.com") // 替换为目标服务器地址
	if err != nil {
		log.Println("Parse URL error:", err)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(WrapConnToResponseWriter(clientConn), &http.Request{})
}

// WrapConnToResponseWriter 将 net.Conn 包装成 http.ResponseWriter
func WrapConnToResponseWriter(conn net.Conn) http.ResponseWriter {
	return &model.ConnResponseWriter{Conn: conn}
}
