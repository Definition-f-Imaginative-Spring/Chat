package main

import (
	"Chat/client/clitool"
	"Chat/server/ConnectManager"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	var conn net.Conn
	var err error
	var success bool

	for {

		conn, err = net.Dial("tcp", "127.0.0.1:8080")
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
			ok := clitool.InputUI(conn, reader)
			if !ok {
				continue
			}
			success = clitool.Register(conn)

		case "2":
			ok := clitool.InputUI(conn, reader)
			if !ok {
				continue
			}
			success = clitool.Login(conn)
		default:
			fmt.Println("无效模式，1=注册，2=登录")
			continue
		}

		if success {
			break
		}
	}

	fmt.Println("已连接到服务器，输入消息后回车即可发送。")
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("conn close err:", err)
		}
	}(conn)

	go clitool.StartHeartbeat(conn, 10*time.Second)

	// 接收消息
	go clitool.Recv(conn)

	// 发送消息
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
