package main

import (
	"Chat/client/clitool"
	"Chat/server/ConnectManager"
	"Chat/server/sertool"
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("主程序 panic: %v\n", err)
		}
	}()

	var conn net.Conn
	var err error
	var success bool
	reader := bufio.NewReader(os.Stdin)
	for {
		conn, err = net.Dial("tcp", "127.0.0.1:8080")
		if err != nil {
			sertool.Close(conn)
			continue
		}
		fmt.Println("先选模式")
		fmt.Println("1=注册，2=登录")
		module, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			sertool.Close(conn)
			continue
		}
		module = strings.TrimSpace(module)
		if err := ConnectManager.SendWithPrefix(conn, module); err != nil {
			fmt.Println("发送用户模式:", err)
			sertool.Close(conn)
			continue
		}

		switch module {
		case "1":
			ok := clitool.InputUI(conn, reader)
			if !ok {
				sertool.Close(conn)
				continue
			}
			success = clitool.Register(conn)

		case "2":
			ok := clitool.InputUI(conn, reader)
			if !ok {
				sertool.Close(conn)
				continue
			}
			success = clitool.Login(conn)
		default:
			fmt.Println("无效模式，1=注册，2=登录")
			sertool.Close(conn)
			continue
		}

		if success {
			break
		}
	}

	fmt.Println("已连接到服务器，输入消息后回车即可发送。")
	fmt.Println("操作手册")
	fmt.Println("输入PAI查看活跃度排名")
	fmt.Println("输入/list获取在线用户列表")
	fmt.Println("输入/exit退出客户端")

	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("conn close err:", err)
		}
	}(conn)

	quitChan := make(chan struct{})
	go clitool.StartHeartbeat(conn, 10*time.Second)

	// 接收消息
	go clitool.Recv(conn, quitChan)

	// 发送消息
	clitool.Send(conn, reader, quitChan)
}
