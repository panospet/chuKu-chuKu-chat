package main

import (
	"chuKu-chuKu-chat/pkg/api"
	"github.com/go-redis/redis/v7"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	api := api.NewApi(rdb)
	api.Run()
}
