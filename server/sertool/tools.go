package sertool

import (
	"Chat/db"
	"Chat/server/ConnectManager"
	"fmt"
	"net"
	"time"
)

// Register 注册
func Register(conn net.Conn, manager *ConnectManager.ConnectManager) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("注册发生 panic: %v\n", err)
		}
	}()

	user, bo := UserCreate(conn)

	if bo {
		fmt.Println("姓名重复")
		err := ConnectManager.SendWithPrefix(conn, "REGISTER_FAIL:用户名已存在")
		if err != nil {
			fmt.Println("发送反馈信息失败")
			return
		}
		return
	}

	err := user.Insert(db.DB)
	if err != nil {
		fmt.Println("插入用户失败:", err)
		err := conn.Close()
		if err != nil {
			fmt.Println("close err:", err)
			return
		}
		return
	}

	if err = ConnectManager.SendWithPrefix(conn, "REGISTER_OK"); err != nil {
		err := conn.Close()
		if err != nil {
			fmt.Println("close err:", err)
			return
		}
		return
	}

	c := ConnectManager.NewConnection(conn)

	manager.AddUser(user.Name, c)

	err = db.InsertRedis(user.Name)
	if err != nil {
		fmt.Println("插入redis失败")
		return
	}

	fmt.Println("新用户连接:", user.Name)

}

// Login 登录
func Login(conn net.Conn, manager *ConnectManager.ConnectManager) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("登录发生 panic: %v\n", err)
		}
	}()

	user, bo := UserCreate(conn)
	ok, err := user.Boolean(db.DB)
	if err != nil {
		fmt.Println(err)
	}

	if bo && ok {
		if err := ConnectManager.SendWithPrefix(conn, "REGISTER_OK"); err != nil {
			err := conn.Close()
			if err != nil {
				fmt.Println("close err:", err)
				return
			}
			return
		}
		c := ConnectManager.NewConnection(conn)

		manager.AddUser(user.Name, c)

		fmt.Println("欢迎:", user.Name)

	} else {
		err := ConnectManager.SendWithPrefix(conn, "密码或用户名错误")
		if err != nil {
			fmt.Println("错误发送信息")
			err := conn.Close()
			if err != nil {
				fmt.Println("close err:", err)
				return
			}
			return
		}
	}
}

// Close  关闭conn
func Close(conn net.Conn) {
	err := conn.Close()
	if err != nil {
		fmt.Println("close conn error:", err)
		return
	}

}

// UserCreate 创建User
func UserCreate(conn net.Conn) (user db.User, bo bool) {
	// 1. 读用户名
	name, err := ConnectManager.ReadMessage(conn)
	if err != nil {
		fmt.Println("name error:", err)
		Close(conn)
		return
	}

	// 2. 读密码
	pwd, err := ConnectManager.ReadMessage(conn)
	if err != nil {
		fmt.Println("pwd error:", err)
		Close(conn)
		return
	}

	user = db.User{
		Name:     name,
		Password: pwd,
	}
	bo, err = user.Exists(db.DB)
	if err != nil {
		fmt.Println(err)
		Close(conn)
		return
	}
	fmt.Println("收到用户名:", name, "密码:", pwd)
	return user, bo
}

// StartServer 开启服务
func StartServer(listener net.Listener) {
	manager := ConnectManager.NewConnectManager()
	manager.StartTimeoutChecker(10*time.Second, 30)
	go manager.StartStreamConsumer()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		module, err := ConnectManager.ReadMessage(conn)
		if err != nil {
			fmt.Println("module error:", err)
			err := conn.Close()
			if err != nil {
				fmt.Println("close err:", err)
				return
			}
			return
		}
		switch module {
		case "1":
			go Register(conn, manager)

		case "2":
			go Login(conn, manager)
		}

	}
}
