package ConnectManager

import (
	"fmt"
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
func (cm *ConnectManager) AddUser(username string, conn *Connection) bool {
	cm.Mutex.Lock()
	defer cm.Mutex.Unlock()

	if _, exists := cm.Connections[username]; exists {
		return false
	}

	conn.Username = username
	cm.Connections[username] = conn

	go cm.ListenRecv(conn)

	cm.Broadcast(username, fmt.Sprintf("%s 上线了", username))

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

	cm.Broadcast(username, fmt.Sprintf("%s 下线了", username))
}

// Broadcast 广播消息给所有用户（除了 sender）
func (cm *ConnectManager) Broadcast(sender string, msg string) {

	for username, conn := range cm.Connections {
		select {
		case conn.SendChan <- fmt.Sprintf("%s: %s", sender, msg):
		default:
			fmt.Println("send buffer full for", username)
		}
	}
}

// SendTo 私聊消息
func (cm *ConnectManager) SendTo(target string, sender string, msg string) {

	conn, ok := cm.Connections[target]
	if !ok {
		fmt.Println("user not online:", target)
		return
	}

	select {
	case conn.SendChan <- fmt.Sprintf("[私聊]%s: %s", sender, msg):
	default:
		fmt.Println("send buffer full for", target)
	}
}

// ListenRecv 监听单个用户的 RecvChan
func (cm *ConnectManager) ListenRecv(conn *Connection) {
	username := conn.Username

	for msg := range conn.RecvChan {
		if msg != "PING" {
			fmt.Println("DEBUG listenRecv 收到:", msg)
		}

		trimmed := strings.TrimSpace(msg)
		if trimmed == "" {
			continue
		}
		if trimmed == "/list" {
			users := cm.ListUsers()
			cm.SendTo(username, "系统", "在线用户: "+strings.Join(users, ", "))
			continue
		}

		if trimmed == "/exit" {
			cm.SendTo(username, "系统", "你已退出")
			cm.RemoveUser(username)
			break
		}

		if trimmed == "PING" {
			conn.LastSeen = time.Now().Unix()
			continue
		}

		if targetUser, content, isPrivate := ParsePrivateMessage(trimmed); isPrivate {
			// 私聊
			_, ok := cm.Connections[targetUser]

			if ok {
				fmt.Printf("用户%s私聊%s：%s\n", username, targetUser, content)
				cm.SendTo(targetUser, username, content)
			} else {
				fmt.Println("不存在该用户")
				cm.SendTo(username, "系统", fmt.Sprintf("用户%s不存在，私聊失败", targetUser))
				continue
			}

		} else {
			fmt.Printf("用户%s:%s\n", username, trimmed)
			cm.Broadcast(username, trimmed)
		}
	}
}

// ParsePrivateMessage 判断是否为私聊
func ParsePrivateMessage(msg string) (string, string, bool) {
	msg = strings.TrimSpace(msg)
	if !strings.HasPrefix(msg, "[private]") || len(msg) <= 9 {
		return "", "", false
	}

	// 去掉 [private] 前缀
	message := msg[9:]
	parts := strings.SplitN(message, ":", 2)
	if len(parts) < 2 {
		return "", "", true
	}

	targetUser := parts[0]
	content := parts[1]
	return targetUser, content, true
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
	ticker := time.NewTicker(interval)
	go func() {
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
