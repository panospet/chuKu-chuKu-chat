package main

import (
	"github.com/go-redis/redis/v7"
	"log"
	"os"

	"chuKu-chuKu-chat/config"
	"chuKu-chuKu-chat/internal/api"
	"chuKu-chuKu-chat/internal/db"
	"chuKu-chuKu-chat/internal/info_fetch"
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

	infoGetter := info_fetch.NewAzuraGetter(cfg.NowPlayingUrl)

	var database db.DbI
	var mode string
	if cfg.Mode == config.DummyMode {
		dummy, err := db.NewDummyDb(rdb)
		if err != nil {
			log.Fatalln("error creating new dummy database:", err)
		}
		database = dummy
		mode = config.DummyMode
	} else if cfg.Mode == config.DbMode {
		postgresDb, err := db.NewPostgresDb(cfg.Dsn, rdb)
		if err != nil {
			log.Fatalln("error creating new database:", err)
		}
		database = postgresDb
		mode = config.DbMode
	} else {
		log.Fatalln("unknown mode")
	}
	app := api.NewApp(mode, rdb, database, infoGetter)
	app.Run()
}
