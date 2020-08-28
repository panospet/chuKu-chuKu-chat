package main

import (
	"flag"
	"github.com/go-redis/redis/v7"
	"os"
)

func main() {
	var input string
	flag.StringVar(&input, "input", "123", "give input")
	flag.Parse()

	var redisAddr string
	redisAddr = os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	rdb.Publish("general", input)
}
