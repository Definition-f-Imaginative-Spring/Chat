package ConnectManager

import (
	"Chat/db"
	"fmt"
	"strings"
	"time"
)

type Msg struct {
	Sender  string
	Types   string
	Content string
}

func NewMsg(sender, types, content string) *Msg {
	return &Msg{Sender: sender, Types: types, Content: content}
}

// SendToStream 消息加入流中
func (m *Msg) SendToStream() {
	if m.Types == "true" {
		_, err := db.AddStreamMessage("chat_stream", m.Sender, m.Content)
		if err != nil {
			fmt.Println("写入 Redis Stream 失败:", err)
		}
	}
}

// Special 特殊消息处理
func (m *Msg) Special(conn *Connection, cm *ConnectManager) (bool, bool) {
	switch m.Content {
	case "PING":
		conn.LastSeen = time.Now().Unix()
		return true, false
	case "/list":
		users := cm.ListUsers()
		cm.SendTo(m.Sender, "系统", "在线用户: "+strings.Join(users, ", "))
		return true, false
	case "/exit":
		cm.SendTo(m.Sender, "系统", "你已退出")
		cm.RemoveUser(m.Sender)
		return true, true
	case "PAI":
		val, err := db.Allures()
		if err != nil {
			fmt.Println("获取活跃度排名错误")
		} else {
			cm.SendTo(m.Sender, "系统", "活跃度排名: ")
			for idx, user := range val {
				x := fmt.Sprintf("第%d名 %s", idx+1, user)
				cm.SendTo(m.Sender, "系统", x)
			}
		}
		return true, false
	default:
		return false, false
	}
}

// ParsePrivateMessage 判断是否为私聊
func (m *Msg) ParsePrivateMessage() (string, string, bool) {
	msg := strings.TrimSpace(m.Content)
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

// MessageHandle 聊天消息处理
func (m *Msg) MessageHandle(cm *ConnectManager) bool {

	if targetUser, content, isPrivate := m.ParsePrivateMessage(); isPrivate {
		// 私聊
		_, ok := cm.Connections[targetUser]
		if ok {
			fmt.Printf("用户%s私聊%s：%s\n", m.Sender, targetUser, content)
			m.SendToStream()
			cm.SendTo(targetUser, m.Sender, content)
		} else {
			fmt.Println("不存在该用户")
			cm.SendTo(m.Sender, "系统", fmt.Sprintf("用户%s不存在，私聊失败", targetUser))
			return false
		}

	} else {
		fmt.Printf("用户%s:%s\n", m.Sender, m.Content)
		m.Types = "true"
		m.SendToStream()

		err := db.IncrementRedis(m.Sender)
		if err != nil {
			fmt.Println("增加活跃度失败")
			return false
		}
	}
	return true
}
