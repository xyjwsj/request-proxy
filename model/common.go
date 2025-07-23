package model

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
)

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
