package ConnectManager

import (
	"Chat/db"
	"fmt"
	"github.com/redis/go-redis/v9"
	"strings"
	"sync"
	"time"
)

type ConnectManager struct {
	Connections map[string]*Connection
	Mutex       sync.Mutex
}

// NewConnectManager 创建 ConnectManager 实例
func NewConnectManager() *ConnectManager {
	return &ConnectManager{
		Connections: make(map[string]*Connection),
	}
}

// AddUser 添加新用户
func (cm *ConnectManager) AddUser(username string, co *Connection) bool {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	if _, exists := cm.Connections[username]; exists {
		return false
	}

	co.Username = username
	cm.Connections[username] = co
	go cm.ListenRecv(co)

	BroadcastSendSystemMsg(fmt.Sprintf("%s 上线了", username), cm)
	BroadcastSendSystemMsg(fmt.Sprintf("以下的直接输出为历史消息"), cm)

	return true
}

// RemoveUser 移除用户
func (cm *ConnectManager) RemoveUser(username string) {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	conn, exists := cm.Connections[username]
	if !exists {
		fmt.Println("用户不存在")
		return
	}
	delete(cm.Connections, username)
	conn.Close()
	BroadcastSendSystemMsg(fmt.Sprintf("%s 下线了", username), cm)
}

// ListenRecv 监听单个用户的 RecvChan
func (cm *ConnectManager) ListenRecv(conn *Connection) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("ListenRecv panic: %v\n", err)
		}
	}()

	username := conn.Username

	for content := range conn.RecvChan {
		if content != "PING" {
			fmt.Println("DEBUG listenRecv 收到:", content)
		}
		//特殊消息处理
		trimmed := strings.TrimSpace(content)
		if trimmed == "" {
			continue
		}

		msg := NewMsg(username, "false", trimmed)

		bool1, bool2 := msg.Special(conn, cm)
		if bool1 && !bool2 {
			continue
		}
		if bool2 {
			break
		}
		//聊天内容处理
		msg.MessageHandle(cm)
		bool3 := msg.Dispatch(cm)
		if !bool3 {
			continue
		}
	}
}

// ListUsers 获取在线用户列表
func (cm *ConnectManager) ListUsers() []string {
	users := make([]string, 0, len(cm.Connections))
	for username := range cm.Connections {
		users = append(users, username)
	}
	return users
}

// StartTimeoutChecker 启动超时检测
func (cm *ConnectManager) StartTimeoutChecker(interval time.Duration, timeoutSec int64) {
	ticker := time.NewTicker(interval) //每隔固定时间发送一个信号
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("TimeoutChecker panic: %v\n", err)
			}
		}()
		for range ticker.C {
			now := time.Now().Unix()
			for username, conn := range cm.Connections {
				if now-conn.LastSeen > timeoutSec {
					fmt.Printf("用户 %s 超时未响应，强制下线\n", username)
					delete(cm.Connections, username)
					conn.Close()
				}
			}

		}
	}()
}

// StartStreamConsumer 消费流中消息
func (cm *ConnectManager) StartStreamConsumer(user *db.User) {
	streamName := "chat_stream"
	lastID := user.LastMessage // 从数据库记录的 last message 开始读

	if lastID == "" {
		lastID = "0"
	}

	go func() {
		for {
			msgs, err := db.ReadStreamByID(streamName, lastID, 10, 5000)
			if err != nil {
				fmt.Println("Stream read error:", err)
				time.Sleep(time.Second)
				continue
			}

			if len(msgs) == 0 {
				continue
			}
			lastID = cm.HandleMessage(user, lastID, msgs)
		}
	}()
}

// HandleMessage 处理消息集
func (cm *ConnectManager) HandleMessage(user *db.User, lastID string, msgs []redis.XMessage) string {
	for _, msg := range msgs {
		sender := msg.Values["sender"].(string)
		content := msg.Values["message"].(string)

		m := &Msg{
			Sender:  sender,
			Types:   "Stream",
			Content: content,
			Target:  user.Name,
		}
		m.Dispatch(cm)

		// 更新 LastMessage
		lastID = msg.ID
		user.LastMessage = lastID
		if err := user.Update(db.DB); err != nil {
			fmt.Println("Update last message error:", err)
		}
	}
	return lastID
}
