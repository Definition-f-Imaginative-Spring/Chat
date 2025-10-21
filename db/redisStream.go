package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var streamCtx = context.Background()

// CreateStreamGroup 创建消费组（如果不存在则自动创建）
func CreateStreamGroup(streamName, groupName string) error {
	err := Rdb.XGroupCreateMkStream(streamCtx, streamName, groupName, "$").Err()
	if err != nil && err.Error() != "BUSY GROUP Consumer Group name already exists" {
		return fmt.Errorf("XGROUP create error: %v", err)
	}
	return nil
}

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

// ReadGroupMessages 从消费组中读取消息（阻塞模式）
func ReadGroupMessages(streamName, groupName, consumerName string, count int64, blockMs int64) ([]redis.XMessage, error) {
	//最多读count条，延时blockMs秒，从最后开始读
	streams, err := Rdb.XReadGroup(streamCtx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, ">"},
		Count:    count,
		Block:    time.Duration(blockMs) * time.Millisecond,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("XREADGROUP error: %v", err)
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}

// AckStreamMessage 确认消息已被消费
func AckStreamMessage(streamName, groupName string, msgID string) error {
	_, err := Rdb.XAck(streamCtx, streamName, groupName, msgID).Result()
	if err != nil {
		return fmt.Errorf("XACK error: %v", err)
	}
	return nil
}

// RecoverPendingMessages 读取未确认消息
func RecoverPendingMessages(streamName, groupName, consumerName string, count int64) ([]redis.XMessage, error) {
	streams, err := Rdb.XReadGroup(streamCtx, &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: consumerName,
		Streams:  []string{streamName, "0"}, // 从最早的未确认开始读
		Count:    count,
		Block:    0,
	}).Result()

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, fmt.Errorf("recover pending error: %v", err)
	}
	if len(streams) == 0 {
		return nil, nil
	}
	return streams[0].Messages, nil
}
