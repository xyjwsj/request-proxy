package proxy

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	brotli "github.com/google/brotli/go/cbrotli"
	"github.com/xyjwsj/request-proxy/model"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	ConnectSuccess = "HTTP/1.1 200 Connection Established\r\n\r\n"
	ConnectFailed  = "HTTP/1.1 502 Bad Gateway\r\n\r\n"
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

	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("Read body error:", err)
		return
	}

	body = interceptorRequest(wrapReq, req, body)

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

	responseBody = interceptorResponse(wrapReq, response, responseBody)

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

func interceptorResponse(wrapReq model.WrapRequest, response *http.Response, responseBody []byte) []byte {
	if wrapReq.OnResponse != nil {
		resData := model.ResponseData{
			ID:     wrapReq.ID,
			Code:   0,
			Header: response.Header,
			Body:   string(responseBody),
		}
		onResponse := wrapReq.OnResponse(resData)
		if onResponse.Code >= 0 {
			response.StatusCode = onResponse.Code
			response.Status = fmt.Sprintf("%d %s", onResponse.Code, http.StatusText(onResponse.Code))
		}
		if onResponse.Header != nil {
			for key, values := range response.Header {
				if len(values) > 0 {
					response.Header.Set(key, values[0])
				}
			}
		}
		if onResponse.Body != "" {
			responseBody = []byte(onResponse.Body)
		}
	}
	return responseBody
}

func interceptorRequest(wrapReq model.WrapRequest, req *http.Request, body []byte) []byte {
	if wrapReq.OnRequest != nil {

		reqData := model.RequestData{
			ID:     wrapReq.ID,
			Url:    req.URL.Path,
			Method: req.Method,
			Header: req.Header,
			Query:  req.URL.Query(),
			Body:   string(body),
		}

		request := wrapReq.OnRequest(reqData)
		if request.Header != nil {
			for key, values := range request.Header {
				if len(values) > 0 {
					req.Header.Set(key, values[0])
				}
			}
		}
		if request.Query != nil {
			queryParams := url.Values(request.Query)
			req.URL.RawQuery = queryParams.Encode()
		}
		if request.Body != "" {
			body = []byte(request.Body)
		}
	}
	return body
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
	_, err = fmt.Fprint(wrapReq.Writer, ConnectSuccess)
	if err != nil {
		log.Println("Write 200 failed:", err)
		_, err = fmt.Fprint(wrapReq.Writer, ConnectFailed)
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
		MinVersion: tls.VersionTLS10, // 支持 TLS 1.0~1.3
		MaxVersion: tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			host := info.ServerName
			cert, err := Cache.GetCertificate(host, "443")
			if err != nil {
				return nil, err
			}
			if cerInfo, ok := cert.(tls.Certificate); ok {
				return &cerInfo, nil
			}
			return nil, errors.New("Certificate Error")
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

	body = interceptorRequest(wrapReq, request, body)

	response, err := transport(request)
	if err != nil {
		log.Println(err.Error())
		return
	}
	responseBody, err := readResponseBody(response.Body, response.Header)
	if err != nil {
		log.Println(err.Error())
		return
	}

	responseBody = interceptorResponse(wrapReq, response, responseBody)

	err = writeCompressedResponse(response, responseBody, wrapReq.Conn)

	//response.Body = io.NopCloser(bytes.NewReader(responseBody))
	//response.Header.Set("Content-Length", strconv.Itoa(len(responseBody)))
	//response.ContentLength = int64(len(responseBody))
	//response.Body = io.NopCloser(bytes.NewReader(responseBody))
	//err = response.Write(wrapReq.Conn)
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

func readResponseBody(body io.ReadCloser, header http.Header) ([]byte, error) {
	// 检查 Content-Encoding 是否为 gzip
	switch header.Get("Content-Encoding") {
	case "gzip":
		gzr, err := gzip.NewReader(body)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()
		return io.ReadAll(gzr)
	case "br":
		brr := brotli.NewReader(body)
		return io.ReadAll(brr)
	default:
		return io.ReadAll(body)
	}
}

func writeCompressedResponse(resp *http.Response, body []byte, w io.Writer) error {
	var out io.Writer = &bytes.Buffer{}
	var err error

	// 判断原始响应是否使用了压缩
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzw := gzip.NewWriter(out)
		_, err = gzw.Write(body)
		if err != nil {
			return err
		}
		err = gzw.Close()
		if err != nil {
			return err
		}
		resp.Header.Set("Content-Encoding", "gzip")

	case "br":
		brw := brotli.NewWriter(out, brotli.WriterOptions{})
		_, err = brw.Write(body)
		if err != nil {
			return err
		}
		err = brw.Close()
		if err != nil {
			return err
		}
		resp.Header.Set("Content-Encoding", "br")

	default:
		out = bytes.NewBuffer(body)
	}

	// 设置新的 Body 和 Content-Length
	resp.Body = io.NopCloser(out.(io.Reader))
	resp.ContentLength = int64(out.(*bytes.Buffer).Len())
	resp.Header.Set("Content-Length", strconv.Itoa(out.(*bytes.Buffer).Len()))
	resp.Header.Del("Transfer-Encoding")

	// 写回客户端
	err = resp.Write(w)
	if err != nil {
		return err
	}

	return nil
}
