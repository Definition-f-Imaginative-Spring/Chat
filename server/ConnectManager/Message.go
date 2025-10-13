package ConnectManager

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// SendWithPrefix 加前缀发送
func SendWithPrefix(conn net.Conn, msg string) error {
	data := []byte(msg)
	length := uint32(len(data))
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, length); err != nil {
		return fmt.Errorf("binary write err: %v", err)
	}

	if _, err := buf.Write(data); err != nil {
		return fmt.Errorf("buffer write err: %v", err)
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("conn write err: %v", err)
	}
	return nil
}

// ReadMessage 读取带前缀的信息
func ReadMessage(conn net.Conn) (string, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lenBuf); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint32(lenBuf)

	msgBuf := make([]byte, length)
	if _, err := io.ReadFull(conn, msgBuf); err != nil {
		return "", err
	}

	return string(msgBuf), nil
}
