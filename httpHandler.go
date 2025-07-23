package proxy

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/xyjwsj/request-proxy/model"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
)

// HandleHTTP 处理 HTTP 请求
func HandleHTTP(wrapReq model.WrapRequest) {

	// 解析 HTTP 请求
	req, err := http.ReadRequest(wrapReq.Reader)
	if err != nil {
		log.Println("Read request error:", err)
		return
	}
	defer req.Body.Close()

	if req.Method == "CONNECT" {
		// 处理 CONNECT 请求（HTTPS 隧道）
		handleCONNECT(wrapReq, req)
		return
	}
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

	req.RequestURI = ""
	req.URL.Host = targetURL.Host
	req.URL.Scheme = targetURL.Scheme
	req.Host = targetURL.Host

	// 自定义 RoundTripper 来拦截请求（可选）
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return wrapReq.Conn, nil
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
	err = resp.Write(wrapReq.Conn)
	if err != nil {
		log.Println("Write response error:", err)
	}
}

func handleCONNECT(wrapReq model.WrapRequest, req *http.Request) {
	host := req.URL.Host

	// 2. 连接到目标服务器（例如：example.com:443）
	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Println("Dial to remote server failed:", err)
		return
	}
	serverConn.Close()

	// 1. 返回 200 Connection established 响应
	_, err = fmt.Fprint(wrapReq.Writer, "HTTP/1.1 200 Connection established\r\n\r\n")
	if err != nil {
		log.Println("Write 200 failed:", err)
		return
	}

	certificate, err := Cache.GetCertificate(req.Host, "443")
	if err != nil {
		log.Println(req.Host + "：获取证书失败：" + err.Error())
		return
	}
	if _, ok := certificate.(tls.Certificate); !ok {
		return
	}
	cert := certificate.(tls.Certificate)
	sslConn := tls.Server(wrapReq.Conn, &tls.Config{
		Certificates: []tls.Certificate{cert},
	})
	// ssl校验
	err = sslConn.Handshake()
	if err != nil {
		log.Println("Handshake error:" + err.Error())
	}
}

func copyData(dst, src net.Conn, errChan chan<- error) {
	_, err := io.Copy(dst, src)
	errChan <- err
}
