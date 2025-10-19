package clitool

import (
	"Chat/server/ConnectManager"
	"Chat/server/sertool"
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

// StartHeartbeat 启动心跳检测
func StartHeartbeat(conn net.Conn, interval time.Duration) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("StartHeartbeat panic: %v\n", err)
		}
	}()

	ticker := time.NewTicker(interval)
	for range ticker.C {
		if err := ConnectManager.SendWithPrefix(conn, "PING"); err != nil {
			fmt.Println("心跳发送失败:", err)
			return
		}
	}
}

// Register 注册
func Register(conn net.Conn) bool {
	// 读取注册结果
	resp, err := ConnectManager.ReadMessage(conn)
	if err != nil {
		fmt.Println("接收注册结果失败:", err)
		sertool.Close(conn)
		return false
	}

	if resp != "REGISTER_OK" {
		fmt.Println("注册失败:", resp)
		sertool.Close(conn)
		return false
	}
	return true

}

// Login 登录
func Login(conn net.Conn) bool {
	resp, err := ConnectManager.ReadMessage(conn)
	if err != nil {
		fmt.Println("接收登录结果失败:", err)
		sertool.Close(conn)
		return false
	}

	if resp != "REGISTER_OK" {
		fmt.Println("登录失败:", resp)
		sertool.Close(conn)
		return false
	}
	return true
}

// Recv 接收消息
func Recv(conn net.Conn) string {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Recv panic: %v\n", err)
		}
	}()

	for {
		msg, err := ConnectManager.ReadMessage(conn)
		if err != nil {
			fmt.Println("服务器断开:", err)
			fmt.Println("请输入/exit退出")
			break
		}
		fmt.Println("[收到] ", msg)
	}
	return ""
}

// Send 消息处理
func Send(conn net.Conn, reader *bufio.Reader) {
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "/exit" {
			fmt.Println("退出客户端。")
			return
		}

		if err := ConnectManager.SendWithPrefix(conn, text); err != nil {
			fmt.Println("发送失败:", err)
			return
		}
	}
}

// InputUI 接收用户名跟密码
func InputUI(conn net.Conn, reader *bufio.Reader) bool {

	fmt.Print("请输入用户名: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("read name err ", err)
		return false
	}
	name = strings.TrimSpace(name)
	if err := ConnectManager.SendWithPrefix(conn, name); err != nil {
		fmt.Println("发送用户名失败:", err)
		err := conn.Close()
		if err != nil {
			fmt.Println("close conn:", err)
			return false
		}
		return false
	}

	fmt.Print("请输入密码: ")
	pwd, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("read pwd err ", err)
		return false
	}
	pwd = strings.TrimSpace(pwd)
	if err := ConnectManager.SendWithPrefix(conn, pwd); err != nil {
		fmt.Println("发送密码失败:", err)
		err := conn.Close()
		if err != nil {
			fmt.Println("close conn:", err)
			return false
		}
		return false
	}

	return true
}
