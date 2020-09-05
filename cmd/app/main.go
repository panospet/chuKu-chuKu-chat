package main

import (
	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/internal/api"
	db2 "chuKu-chuKu-chat/internal/db"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	db, err := db2.NewDummyOperations(rdb)
	if err != nil {
		panic(err)
	}
	app := api.NewApp(rdb, db)
	app.Run()
}
