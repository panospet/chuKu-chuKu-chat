package config

import (
	"errors"
	"fmt"
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
	godotenv.Load()
}

func NewConfig() (*Config, error) {
	mode, ok := os.LookupEnv("MODE")
	if !ok {
		return nil, errors.New("MODE env variable does not exist")
	}
	if mode != DummyMode && mode != DbMode {
		return nil, errors.New(fmt.Sprintf("MODE should have either value %s or %s", DummyMode, DbMode))
	}
	dsn := ""
	if mode == DbMode {
		dsn, ok = os.LookupEnv("DSN")
		if !ok {
			return nil, errors.New("DSN env variable does not exist")
		}
	}
	redis, ok := os.LookupEnv("REDIS_URL")
	if !ok {
		return nil, errors.New("REDIS_URL env variable does not exist")
	}
	nowPlayingUrl, ok := os.LookupEnv("NOW_PLAYING_URL")
	if !ok {
		return nil, errors.New("NOW_PLAYING_URL env variable does not exist")
	}
	return &Config{
		Mode:          mode,
		Dsn:           dsn,
		Redis:         redis,
		NowPlayingUrl: nowPlayingUrl,
	}, nil
}
