package model

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"log"
)

type User struct {
	Username         string
	pubSub           *redis.PubSub
	StopListenerChan chan struct{}
	//listening        bool
	MessageChan chan redis.Message
	channels    []string
}

func NewUser(username string) *User {
	return &User{
		Username:         username,
		MessageChan:      make(chan redis.Message),
		StopListenerChan: make(chan struct{}),
		channels:         []string{"general"},
	}
}

func (u *User) GetChannels() ([]string, error) {
	// todo come from db
	return u.channels, nil
}

func (u *User) SubscribeToChannel(channelName string) error {
	// todo come from db
	u.channels = append(u.channels, channelName)
	return nil
}

func (u *User) Connect(rdb *redis.Client) error {
	if _, err := rdb.SAdd("useridia", u.Username).Result(); err != nil {
		return err
	}

	var c []string
	c, _ = u.GetChannels()

	if len(c) == 0 {
		return errors.New(fmt.Sprintf("no channels for user %s", u.Username))
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
	fmt.Println("subscribed to pubsub for channels", c)

	go func() {
		fmt.Println("started listening to pubsub channels")
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

	//go func() {
	//	fmt.Println("listening to received messages")
	//	for msg := range u.MessageChan {
	//		fmt.Println("user", u.Username, "received message", msg)
	//	}
	//}()

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
