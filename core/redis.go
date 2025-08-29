package core

import (
	"context"
	"github.com/go-redis/redis"
	"time"
)

func InitRedis(addr string, pwd string, db int) (client *redis.Client) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
		PoolSize: 100, //连接池的大小
	})
	_, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	_, err := rdb.Ping().Result()
	if err != nil {
		panic(err)
		return
	}
	return rdb
}
