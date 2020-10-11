package main

import (
	"log"
	"os"

	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/config"
	"chuKu-chuKu-chat/internal/api"
	"chuKu-chuKu-chat/internal/db"
)

func main() {
	cfgPath := os.Getenv("CONFIG_FILE")
	if cfgPath == "" {
		cfgPath = "./config/config.yml"
	}
	cfg, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatalln("error creating new configuration:", err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis,
		Password: "",
		DB:       0,
	})

	var database db.DbI
	if cfg.Mode == config.DummyMode {
		dummy, err := db.NewDummyDb(rdb)
		if err != nil {
			log.Fatalln("error creating new dummy database:", err)
		}
		database = dummy
	} else if cfg.Mode == config.DbMode {
		postgresDb, err := db.NewPostgresDb(cfg.Dsn, rdb)
		if err != nil {
			log.Fatalln("error creating new database:", err)
		}
		database = postgresDb
	} else {
		log.Fatalln("unknown mode")
	}
	app := api.NewApp(rdb, database)
	app.Run()
}
