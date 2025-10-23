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

// ReadStreamByID 从流中读取
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
