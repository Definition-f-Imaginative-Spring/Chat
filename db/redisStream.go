package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var streamCtx = context.Background()

// AddStreamMessage 往指定流中添加消息
func AddStreamMessage(streamName string, sender string, msg string) (string, error) {
	id, err := Rdb.XAdd(streamCtx, &redis.XAddArgs{
		Stream: streamName,
		Values: map[string]interface{}{
			"sender":  sender,
			"message": msg,
			"time":    time.Now().Format(time.RFC3339),
		},
	}).Result()
	if err != nil {
		return "", fmt.Errorf("XADD error: %v", err)
	}
	return id, nil
}

// ReadStreamByID 从流中读取历史消息
func ReadStreamByID(streamName string, lastID string, count int64, blockMs int64) ([]redis.XMessage, error) {
	if lastID == "" {
		lastID = "0" // 从最早消息开始读
	}
	streams, err := Rdb.XRead(streamCtx, &redis.XReadArgs{
		Streams: []string{streamName, lastID},
		Count:   count,
		Block:   time.Duration(blockMs) * time.Millisecond, // 阻塞等待新消息
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("XREAD error: %v", err)
	}
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// AutoTrimStream 定期检查并修剪流，保留最多 maxLen 条消息
func AutoTrimStream(streamName string, maxLen int64, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			// XTrim 可以直接自动删除最老的消息
			_, err := Rdb.XTrimMaxLen(streamCtx, streamName, maxLen).Result()
			if err != nil {
				fmt.Printf("Failed to trim stream %s: %v\n", streamName, err)
			}
		}
	}()
}

// GetLatestStreamID 返回流中最新消息的 ID
func GetLatestStreamID(streamName string) (string, error) {
	msgs, err := Rdb.XRevRangeN(streamCtx, streamName, "+", "-", 1).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get latest message: %v", err)
	}

	if len(msgs) == 0 {
		return "0-0", nil // 流为空
	}
	return msgs[0].ID, nil
}
