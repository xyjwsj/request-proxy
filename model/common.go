package model

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

type RequestData struct {
	ID       string              `json:"ID"`
	Host     string              `json:"host"`
	ClientIp string              `json:"clientIp"`
	TargetIp string              `json:"targetIp"`
	Url      string              `json:"url"`
	Method   string              `json:"method"`
	Header   map[string][]string `json:"header"`
	Query    map[string][]string `json:"query"`
	Body     string              `json:"body"`
}

type ResponseData struct {
	ID       string              `json:"ID"`
	Code     int                 `json:"code"`
	Header   map[string][]string `json:"header"`
	Body     string              `json:"body"`
	Duration int64               `json:"duration"`
}

type RequestCall func(data RequestData) RequestData
type ResponseCall func(data ResponseData) ResponseData

type WrapWriter struct {
	io.Writer
}

type WrapRequest struct {
	ID         string
	Conn       net.Conn
	Writer     *bufio.Writer
	Reader     *bufio.Reader
	OnRequest  RequestCall
	OnResponse ResponseCall
	Https      bool
	Duration   int64
}

type ConnResponseWriter struct {
	Conn net.Conn
	Code int
	Hdr  http.Header
}

func (w *ConnResponseWriter) Header() http.Header {
	if w.Hdr == nil {
		w.Hdr = make(http.Header)
	}
	return w.Hdr
}

func (w *ConnResponseWriter) WriteHeader(code int) {
	w.Code = code
}

func (w *ConnResponseWriter) Write(p []byte) (int, error) {
	// 构造 HTTP 响应头
	if w.Hdr == nil {
		w.WriteHeader(http.StatusOK)
	}

	// 拼接响应头和内容
	resp := fmt.Sprintf("HTTP/1.1 %d OK\r\n%s\r\nContent-Length: %d\r\n\r\n%s",
		w.Code,
		w.Hdr,
		len(p),
		p,
	)

	return w.Conn.Write([]byte(resp))
}

type DataReader struct {
	Reader      io.Reader
	InitialData []byte
	once        sync.Once
}

func (r *DataReader) Read(p []byte) (n int, err error) {
	r.once.Do(func() {
		n = copy(p, r.InitialData)
		if n < len(r.InitialData) {
			copy(p[n:], r.InitialData[n:])
			n = len(r.InitialData)
		}
	})
	if n == 0 {
		return r.Reader.Read(p)
	}
	return n, nil
}
