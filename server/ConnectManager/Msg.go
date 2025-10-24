package ConnectManager

import (
	"Chat/db"
	"fmt"
	"strings"
	"time"
)

type Msg struct {
	Sender  string
	Target  string
	Types   string
	Content string
}

// PrivateSendSystemMsg 系统私聊
func PrivateSendSystemMsg(target, content string, cm *ConnectManager) {
	msg := Msg{
		Sender:  "系统",
		Content: content,
	}
	msg.SendTo(target, cm)
}

// BroadcastSendSystemMsg 系统广播
func BroadcastSendSystemMsg(content string, cm *ConnectManager) {
	msg := Msg{
		Sender:  "系统",
		Content: content,
	}
	msg.Broadcast(cm)
}

func NewMsg(sender, types, content string) *Msg {
	return &Msg{Sender: sender, Types: types, Content: content}
}

// SendToStream 消息加入流中
func (m *Msg) SendToStream() {
	_, err := db.AddStreamMessage("chat_stream", m.Sender, m.Content)
	if err != nil {
		fmt.Println("写入 Redis Stream 失败:", err)
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
		PrivateSendSystemMsg(m.Sender, "在线用户: "+strings.Join(users, ", "), cm)
		return true, false
	case "/exit":
		PrivateSendSystemMsg(m.Sender, "你已退出", cm)
		cm.RemoveUser(m.Sender)
		return true, true
	case "PAI":
		val, err := db.Allures()
		if err != nil {
			fmt.Println("获取活跃度排名错误")
		} else {
			PrivateSendSystemMsg(m.Sender, "活跃度排名: ", cm)
			for idx, user := range val {
				x := fmt.Sprintf("第%d名 %s", idx+1, user)
				PrivateSendSystemMsg(m.Sender, x, cm)
			}
		}
		return true, false
	default:
		return false, false
	}
}

// ParsePrivateMessage 判断是否为私聊
func (m *Msg) ParsePrivateMessage() (string, string, bool, bool) {
	msg := strings.TrimSpace(m.Content)
	if !strings.HasPrefix(msg, "[private]") || len(msg) <= 9 {
		return "", "", false, false
	}
	// 去掉 [private] 前缀
	message := msg[9:]
	parts := strings.SplitN(message, ":", 2)
	if len(parts) < 2 {
		return "", "", true, false
	}

	targetUser := parts[0]
	content := parts[1]
	return targetUser, content, true, true
}

// MessageHandle 确定消息类型
func (m *Msg) MessageHandle(cm *ConnectManager) {
	targetUser, content, isPrivate, isTrue := m.ParsePrivateMessage()

	if isPrivate && isTrue {
		m.Types = "private"
		m.Content = content
		m.Target = targetUser
	}

	if isPrivate && !isTrue {
		PrivateSendSystemMsg(m.Sender, "私聊格式错误", cm)
		m.Types = ""
	}

	if !isPrivate {
		m.Types = "broadcast"
	}
}

// Dispatch 消息处理
func (m *Msg) Dispatch(cm *ConnectManager) bool {
	switch m.Types {
	case "private":
		return m.HandlePrivate(cm)
	case "broadcast":
		return m.HandleBroadcast()
	case "Stream":
		return m.HandleStream(cm)
	default:
		fmt.Println("未知消息类型:", m.Types)
		return false
	}
}

// HandleBroadcast 处理广播
func (m *Msg) HandleBroadcast() bool {
	fmt.Printf("用户%s:%s\n", m.Sender, m.Content)
	m.SendToStream()
	err := db.IncrementRedis(m.Sender)
	if err != nil {
		fmt.Println("增加活跃度失败")
		return false
	}
	return true
}

// HandlePrivate 处理私聊
func (m *Msg) HandlePrivate(cm *ConnectManager) bool {
	_, ok := cm.Connections[m.Target]
	if ok {
		fmt.Printf("用户%s私聊%s：%s\n", m.Sender, m.Target, m.Content)
		m.SendTo(m.Target, cm)
	} else {
		fmt.Println("不存在该用户")
		PrivateSendSystemMsg(m.Sender, "不存在该用户", cm)
		return false
	}
	return true
}

// Broadcast 广播消息给所有用户
func (m *Msg) Broadcast(cm *ConnectManager) {
	for username, conn := range cm.Connections {
		select {
		case conn.SendChan <- fmt.Sprintf("%s: %s", m.Sender, m.Content):
		default:
			fmt.Println("send buffer full for", username)
		}
	}
}

// SendTo 私聊消息
func (m *Msg) SendTo(target string, cm *ConnectManager) {
	conn, ok := cm.Connections[target]
	if !ok {
		fmt.Println("user not online:", target)
		return
	}
	select {
	case conn.SendChan <- fmt.Sprintf("[私聊]%s: %s", m.Sender, m.Content):
	default:
		fmt.Println("send buffer full for", target)
	}
}

// Stream 只发送自己
func (m *Msg) Stream(target string, cm *ConnectManager) {
	conn, ok := cm.Connections[target]
	if !ok {
		fmt.Println("user not online:", target)
		return
	}
	select {
	case conn.SendChan <- fmt.Sprintf("%s: %s", m.Sender, m.Content):
	default:
		fmt.Println("send buffer full for", target)
	}
}

// HandleStream 处理stream
func (m *Msg) HandleStream(cm *ConnectManager) bool {
	m.Stream(m.Target, cm)
	return true
}
