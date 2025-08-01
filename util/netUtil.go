package util

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// WriteFullResponse 手动构建并写入HTTP响应
func WriteFullResponse(conn io.Writer, resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}

	// 确保必要的HTTP协议字段存在
	if resp.Proto == "" {
		resp.Proto = "HTTP/1.1"
	}
	if resp.ProtoMajor == 0 && resp.ProtoMinor == 0 {
		resp.ProtoMajor = 1
		resp.ProtoMinor = 1
	}

	// 写入状态行: HTTP/1.1 200 OK\r\n
	statusLine := fmt.Sprintf("%s %d %s\r\n", resp.Proto, resp.StatusCode, http.StatusText(resp.StatusCode))
	_, err := conn.Write([]byte(statusLine))
	if err != nil {
		return fmt.Errorf("failed to write status line: %w", err)
	}

	// 写入响应头
	for key, values := range resp.Header {
		for _, value := range values {
			headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
			_, err := conn.Write([]byte(headerLine))
			if err != nil {
				return fmt.Errorf("failed to write header %s: %w", key, err)
			}
		}
	}

	// 写入空行分隔头部和响应体
	_, err = conn.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write header/body separator: %w", err)
	}

	// 写入响应体
	if resp.Body != nil {
		_, err = io.Copy(conn, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write response body: %w", err)
		}
	}

	return nil
}

// GetRealClientIP 更完善的获取客户端IP函数
func GetRealClientIP(req *http.Request) string {
	// 按优先级尝试各种方式获取IP

	// 1. 尝试 X-Forwarded-For (最常用)
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				return ip
			}
		}
		// 如果都是私有IP，返回第一个
		return strings.TrimSpace(ips[0])
	}

	// 2. 尝试 X-Real-IP
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 3. 尝试其他代理头部
	if xff := req.Header.Get("X-Forwarded"); xff != "" {
		return xff
	}

	if xfhost := req.Header.Get("X-Forwarded-Host"); xfhost != "" {
		return xfhost
	}

	if xci := req.Header.Get("X-Client-IP"); xci != "" {
		return xci
	}

	// 4. 最后从连接获取直接客户端IP
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return req.RemoteAddr
	}
	return host
}

func GetClientIP(conn net.Conn) string {
	if conn == nil {
		return ""
	}
	addr := conn.RemoteAddr().String()
	if addr == "" {
		return ""
	}
	split := strings.Split(addr, ":")
	if len(split) == 1 {
		return addr
	}
	return split[0]
}

func GetIPFromDomain(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}
	return ipStrings, nil
}
