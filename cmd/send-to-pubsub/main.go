package main

import (
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/internal/api"
	"chuKu-chuKu-chat/internal/db"
)

func main() {
	var input string
	var channel string
	var redisAddr string
	flag.StringVar(&input, "input", "test message", "give input")
	flag.StringVar(&channel, "channel", "general", "channel post")
	flag.StringVar(&redisAddr, "redis", "localhost:6379", "redis address")
	flag.Parse()

	if input == "" {
		log.Fatalf("message cannot be empty")
	}
	if channel == "" {
		log.Fatalf("please give a valid channel")
	}

	p := api.Payload{
		Content:   input,
		Channel:   "general",
		User:      db.KickItBotUsername,
		UserColor: "#000000",
		Command:   2,
		Timestamp: time.Now(),
	}
	b, _ := json.Marshal(p)

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	rdb.Publish(channel, string(b))
}
