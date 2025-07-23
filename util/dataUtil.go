package util

import (
	"io"
	"log"
	"net"
)

// CopyData 数据复制函数
func CopyData(dst, src net.Conn, done chan<- struct{}) {
	_, err := io.Copy(dst, src)
	if err != nil {
		log.Println("Copy error:", err)
	}
	done <- struct{}{}
}
