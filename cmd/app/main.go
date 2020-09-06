package main

import (
	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/internal/api"
	"chuKu-chuKu-chat/internal/operations"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	dummOp, err := operations.NewDummyOperator(rdb)
	if err != nil {
		panic(err)
	}
	app := api.NewApp(rdb, dummOp)
	app.Run()
}
