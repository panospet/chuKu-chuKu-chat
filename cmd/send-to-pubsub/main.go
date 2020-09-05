package main

import (
	"flag"
	"github.com/go-redis/redis/v7"
	"log"
	"os"
)

func main() {
	var input string
	var channel string
	flag.StringVar(&input, "input", "{\"content\":\"asdf\",\"channel\":\"general\",\"command\":2,\"user\":\"admin\",\"timestamp\":\"2020-09-05T23:48:23.553793195+03:00\"}", "give input")
	flag.StringVar(&channel, "channel", "general", "channel post")
	flag.Parse()

	if input == "" {
		log.Fatalf("message cannot be empty")
	}
	if channel == "" {
		log.Fatalf("please give a valid channel")
	}

	var redisAddr string
	redisAddr = os.Getenv("REDIS_ADDRESS")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	rdb.Publish(channel, input)
}
