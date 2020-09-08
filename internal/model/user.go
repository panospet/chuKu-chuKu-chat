package model

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
	"time"
)

type User struct {
	Id        int       `json:"id" db:"id"`
	Username  string    `json:"name" db:"name"`
	Channels  []string  `json:"channels"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	pubSub           *redis.PubSub      `json:"-"`
	StopListenerChan chan struct{}      `json:"-"`
	MessageChan      chan redis.Message `json:"-"`
}

func NewUser(username string, channels ...string) *User {
	generalExists := false
	for _, ch := range channels {
		if ch == "general" {
			generalExists = true
		}
	}
	if !generalExists {
		channels = append(channels, "general")
	}
	return &User{
		Username:         username,
		MessageChan:      make(chan redis.Message),
		StopListenerChan: make(chan struct{}),
		Channels:         channels,
		CreatedAt:        time.Now(),
	}
}

func (u *User) GetChannels() []string {
	return u.Channels
}

func (u *User) AddChannel(channelName string) bool {
	alreadyExists := false
	for _, c := range u.Channels {
		if c == channelName {
			alreadyExists = true
			break
		}
	}
	if !alreadyExists {
		u.Channels = append(u.Channels, channelName)
	}
	return alreadyExists
}

func (u *User) RefreshChannels(rdb *redis.Client) error {
	if err := u.ConnectToPubSub(rdb); err != nil {
		return errors.New(fmt.Sprintf("error during user connection: %s", err))
	}
	return nil
}

// todo violation; needs to be moved elsewhere
func (u *User) ConnectToPubSub(rdb *redis.Client) error {
	if _, err := rdb.SAdd("useridia", u.Username).Result(); err != nil {
		return err
	}

	c := u.GetChannels()
	if len(c) == 0 {
		return errors.New(fmt.Sprintf("no Channels for user %s", u.Username))
	}

	// if user has already a pubsub instance it needs to be closed
	if u.pubSub != nil {
		if err := u.pubSub.Unsubscribe(); err != nil {
			return errors.New(fmt.Sprintf("error unsubscribing from pubsub: %s", err))
		}
		if err := u.pubSub.Close(); err != nil {
			return errors.New(fmt.Sprintf("error closing pubsub connection: %s", err))
		}
	}

	u.pubSub = rdb.Subscribe(c...)
	fmt.Println("user", u.Username, "subscribed to pubsub for Channels", c)

	go func() {
		fmt.Println("started listening to pubsub Channels")
		for {
			select {
			case msg, ok := <-u.pubSub.Channel():
				if !ok {
					log.Println("something bad happened, terminating pubsub listener")
					return
				}
				u.MessageChan <- *msg
			case <-u.StopListenerChan:
				fmt.Println("listening to pubpub stopped.")
				return
			}
		}
	}()

	return nil
}

func (u *User) Disconnect() error {
	if u.pubSub != nil {
		if err := u.pubSub.Unsubscribe(); err != nil {
			return err
		}
		if err := u.pubSub.Close(); err != nil {
			return err
		}
	}
	u.StopListenerChan <- struct{}{}
	close(u.MessageChan)

	return nil
}
