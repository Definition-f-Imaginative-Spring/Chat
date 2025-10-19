package ConnectManager

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ConnMsg struct {
	Conn     *Connection
	Username string
	Content  string
}

type Connection struct {
	Conn     net.Conn
	RecvChan chan string
	SendChan chan string
	QuitChan chan struct{}
	Username string
	LastSeen int64
	once     sync.Once
}

func NewConnection(conn net.Conn) *Connection {
	c := &Connection{
		Conn:     conn,
		RecvChan: make(chan string, 20),
		SendChan: make(chan string, 20),
		QuitChan: make(chan struct{}),
		LastSeen: time.Now().Unix(),
	}
	go c.ReadLoop()
	go c.WriteLoop()

	return c
}

func (c *Connection) ReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ReadLoop panic: %v\n", err)
		}
		fmt.Println("ReadLoop exit")
	}()

	for {
		msg, err := ReadMessage(c.Conn)
		if err != nil {
			fmt.Println("read error:", err)
			c.Close()
			return
		}

		select {
		case c.RecvChan <- msg:
		case <-c.QuitChan:
			return
		}
	}
}

func (c *Connection) WriteLoop() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("WriteLoop panic: %v\n", err)
		}
		fmt.Println("WriteLoop exit")
	}()

	for {
		select {
		case <-c.QuitChan:
			return
		case msg, ok := <-c.SendChan:
			if !ok {
				return
			}
			if err := SendWithPrefix(c.Conn, msg); err != nil {
				fmt.Println("write error:", err)
				c.Close()
				return
			}
		}
	}
}

func (c *Connection) Close() {
	c.once.Do(func() {
		close(c.QuitChan)
		err := c.Conn.Close()
		if err != nil {
			fmt.Println("close conn error:", err)
			return
		}
	})
}
