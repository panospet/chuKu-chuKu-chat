package main

import (
	"fmt"
	"log"

	"github.com/go-redis/redis/v7"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping().Result()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(pong)
}
