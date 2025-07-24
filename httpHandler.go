package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/xyjwsj/request-proxy/model"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
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

	response, err := transport(req)

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(string(responseBody))

	response.Body = io.NopCloser(bytes.NewReader(responseBody))
	response.Header.Set("Content-Length", strconv.Itoa(len(responseBody)))
	response.ContentLength = int64(len(responseBody))
	response.Body = io.NopCloser(bytes.NewReader(responseBody))
	err = response.Write(wrapReq.Conn)
	if err != nil {
		log.Println(err.Error())
		return
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
	_ = wrapReq.Writer.Flush() // 立即刷新 buffer，确保响应已发送

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
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			host := req.Host
			cert, err := Cache.GetCertificate(host, "443")
			if err != nil {
				return nil, err
			}
			return cert.(*tls.Certificate), nil
		},
	})
	// ssl校验
	err = sslConn.Handshake()
	if err != nil {
		log.Println("Handshake error:" + err.Error())
	}

	wrapReq.Conn = sslConn
	wrapReq.Reader = bufio.NewReader(wrapReq.Conn)
	wrapReq.Writer = bufio.NewWriter(wrapReq.Conn)
	request, err := http.ReadRequest(wrapReq.Reader)
	if err != nil {
		log.Println(err.Error())
		return
	}

	body, _ := io.ReadAll(request.Body)
	log.Println(string(body))
	request.Body = io.NopCloser(bytes.NewReader(body))
	request = setRequest(request)
	response, err := transport(request)
	if err != nil {
		log.Println(err.Error())
		return
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println(string(responseBody))

	response.Body = io.NopCloser(bytes.NewReader(responseBody))
	response.Header.Set("Content-Length", strconv.Itoa(len(responseBody)))
	response.ContentLength = int64(len(responseBody))
	response.Body = io.NopCloser(bytes.NewReader(responseBody))
	err = response.Write(wrapReq.Conn)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("END")
}

func transport(request *http.Request) (*http.Response, error) {
	// 去除一些头部
	response, err := (&http.Transport{
		DisableKeepAlives:     true,
		ResponseHeaderTimeout: 60 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}).RoundTrip(request)
	if err != nil {
		return nil, err
	}
	// 去除一些头部
	return response, err
}

func setRequest(request *http.Request) *http.Request {
	request.Header.Set("Connection", "false")
	request.URL.Host = request.Host
	request.URL.Scheme = "https"
	return request
}
