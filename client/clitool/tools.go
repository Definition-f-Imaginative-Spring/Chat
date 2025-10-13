package clitool

import (
	"Chat/server/ConnectManager"
	"Chat/server/sertool"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

// StartHeartbeat 启动心跳检测
func StartHeartbeat(conn net.Conn, interval time.Duration) {
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
	for {
		msg, err := ConnectManager.ReadMessage(conn)
		if err != nil {
			fmt.Println("服务器断开:", err)
			os.Exit(0)
		}
		fmt.Println("[收到] ", msg)
	}
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

func StartClient(success bool, reader *bufio.Reader) {
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			panic(err)
		}
		fmt.Println("先选模式")
		fmt.Println("1=注册，2=登录")
		module, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		module = strings.TrimSpace(module)
		if err := ConnectManager.SendWithPrefix(conn, module); err != nil {
			fmt.Println("发送用户模式:", err)
			err := conn.Close()
			if err != nil {
				fmt.Println("close err:", err)
				return
			}
			continue
		}

		switch module {
		case "1":
			ok := InputUI(conn, reader)
			if !ok {
				continue
			}
			success = Register(conn)

		case "2":
			ok := InputUI(conn, reader)
			if !ok {
				continue
			}
			success = Login(conn)
		default:
			fmt.Println("无效模式，1=注册，2=登录")
			continue
		}

		if success {
			break
		}
	}
}
