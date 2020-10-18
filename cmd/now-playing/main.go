package main

import (
	"chuKu-chuKu-chat/config"
	"chuKu-chuKu-chat/internal/info_fetch"
	"encoding/json"
	"fmt"
	"log"
)

func main() {
	cfg, err := config.NewConfig("./config/config.yml")
	if err != nil {
		log.Fatalln("error creating new configuration:", err)
	}
	infoGetter := info_fetch.NewAzuraGetter(cfg.NowPlayingUrl)
	info, err := infoGetter.Get()
	if err != nil {
		log.Fatalln(err)
	}
	infoBytes, err := json.Marshal(info)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(infoBytes))
}
