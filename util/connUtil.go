package util

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"request-proxy/model"
)

// IsHTTPRequest 判断是否为 HTTP 请求
func IsHTTPRequest(data []byte) bool {
	return len(data) > 4 && string(data[:4]) == "GET "
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
