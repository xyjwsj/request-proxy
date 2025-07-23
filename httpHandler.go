package proxy

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"request-proxy/model"
)

// HandleHTTP 处理 HTTP 请求
func HandleHTTP(clientConn net.Conn, data []byte) {
	// 创建缓冲 reader，合并已读取的部分数据和连接本身
	reader := bufio.NewReader(&model.DataReader{
		Reader:      clientConn,
		InitialData: data,
	})

	// 解析 HTTP 请求
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Println("Read request error:", err)
		return
	}
	defer req.Body.Close()

	// 你可以在这里修改请求头或 body
	// ----------------------------------
	// 修改 Header 示例
	//req.Header.Set("X-Forwarded-By", "MyProxy")

	// 如果需要删除某些 Header
	//req.Header.Del("User-Agent")
	// TODO 修改请求头

	// 如果需要读取并修改 Body，请参考下方 Body 处理部分
	// 读取 body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("Read body error:", err)
		return
	}

	// 修改 body 内容
	// TODO 修改请求Body
	//body = bytes.Replace(body, []byte("old"), []byte("new"), -1)

	// 重新设置 Body
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	req.ContentLength = int64(len(body))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(body)))
	// ----------------------------------

	// 设置目标地址
	// 自动获取目标地址
	targetHost := req.Host
	if targetHost == "" {
		targetHost = req.URL.Host
	}

	if targetHost == "" {
		log.Println("Cannot determine target host")
		return
	}

	scheme := "http"
	if req.TLS != nil || req.URL.Scheme == "https" {
		scheme = "https"
	}

	targetURL, err := url.Parse(fmt.Sprintf("%s://%s", scheme, targetHost))
	if err != nil {
		log.Println("Parse URL error:", err)
		return
	}

	req.URL = targetURL

	// 自定义 RoundTripper 来拦截请求（可选）
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return clientConn, nil
		},
	}

	// 使用自定义 Transport 的 Client 发送请求
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 记录代理请求
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Client Do error:", err)
		return
	}
	defer resp.Body.Close()

	// 将响应写回客户端
	err = resp.Write(clientConn)
	if err != nil {
		log.Println("Write response error:", err)
	}
}
