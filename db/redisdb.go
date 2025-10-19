package db

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

// Init 初始化redis
func init() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	if err := Rdb.Ping(ctx).Err(); err != nil {
		fmt.Println("连接Redis失败")
	}
}

// InsertRedis 将用户加入member中
func InsertRedis(username string) error {
	ctx := context.Background()
	err := Rdb.ZAdd(ctx, "key", redis.Z{Score: 0, Member: username}).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// IncrementRedis 指定用户活跃度增加
func IncrementRedis(username string) error {
	ctx := context.Background()
	err := Rdb.ZIncrBy(ctx, "key", 1, username).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// Allures 显示排名
func Allures() ([]string, error) {
	ctx := context.Background()
	val, err := Rdb.ZRevRange(ctx, "key", 0, -1).Result()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return val, nil
}
