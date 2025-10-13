package main

import (
	"Chat/db"
	"Chat/server/sertool"
	"fmt"
	"net"
)

func main() {

	err := db.InitDB()
	if err != nil {
		fmt.Println("连接数据库错误")
		return
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener close err:", err)
		}
	}(listener)

	fmt.Println("服务器已启动，监听端口 8080 ...")

	sertool.StartServer(listener)

}
