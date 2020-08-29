package main

import (
	"chuKu-chuKu-chat/pkg/api"
	db2 "chuKu-chuKu-chat/pkg/db"
	"github.com/go-redis/redis/v7"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	db := db2.NewDummyDb()
	api := api.NewApi(rdb, db)
	api.Run()
}
