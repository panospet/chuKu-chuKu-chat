package main

import (
	"log"

	"github.com/go-redis/redis/v7"

	"chuku-chuku-chat/config"
	"chuku-chuku-chat/internal/api"
	"chuku-chuku-chat/internal/db"
	"chuku-chuku-chat/internal/info_fetch"
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
