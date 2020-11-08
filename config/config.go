package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const DummyMode = "dummy"
const DbMode = "db"

type Config struct {
	Mode          string
	Dsn           string
	Redis         string
	NowPlayingUrl string
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln("No .env file found")
	}
}

func NewConfig() (*Config, error) {
	mode, ok := os.LookupEnv("MODE")
	if !ok {
		return nil, errors.New("MODE does not exist in .env file")
	}
	if mode != DummyMode && mode != DbMode {
		return nil, errors.New(fmt.Sprintf("MODE should have either value %s or %s", DummyMode, DbMode))
	}
	dsn := ""
	if mode == DbMode {
		dsn, ok = os.LookupEnv("DSN")
		if !ok {
			return nil, errors.New("DSN does not exist in .env file")
		}
	}
	redis, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		return nil, errors.New("REDIS_URL does not exist in .env file")
	}
	nowPlayingUrl, ok := os.LookupEnv("NOW_PLAYING_URL")
	if !ok {
		return nil, errors.New("NOW_PLAYING_URL does not exist in .env file")
	}
	return &Config{
		Mode:          mode,
		Dsn:           dsn,
		Redis:         redis,
		NowPlayingUrl: nowPlayingUrl,
	}, nil
}
