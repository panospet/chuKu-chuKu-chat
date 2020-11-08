package main

import (
	"log"

	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/config"
	"chuKu-chuKu-chat/internal/common"
	"chuKu-chuKu-chat/internal/db"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalln("error creating new configuration:", err)
	}
	optRedis, err := redis.ParseURL(cfg.Redis)
	if err != nil {
		log.Fatalln("redis url could not be parsed")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     optRedis.Addr,
		Password: optRedis.Password,
		DB:       optRedis.DB,
	})

	postgresDb, err := db.NewPostgresDb(cfg.Dsn, rdb)
	if err != nil {
		log.Fatalln("error creating new database:", err)
	}

	users, err := postgresDb.GetUsers()
	if err != nil {
		panic(err)
	}
	var usernames []string
	for _, u := range users {
		usernames = append(usernames, u.Username)
	}

	channels, err := postgresDb.GetChannels()
	if err != nil {
		panic(err)
	}

	for _, c := range channels {
		messages := common.GenerateRandomMessages(c.Name, 100, usernames...)
		for _, msg := range messages {
			if err := postgresDb.AddMessage(msg); err != nil {
				panic(err)
			}
		}
	}
}
