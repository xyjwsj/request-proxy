package util

import (
	"fmt"
	"io"
	"net/http"
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
