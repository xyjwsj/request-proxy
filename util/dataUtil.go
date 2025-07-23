package util

import (
	"io"
	"net"
)

// CopyData 数据复制函数
func CopyData(dst, src net.Conn, errChan chan<- error) {
	_, err := io.Copy(dst, src)
	errChan <- err
}
