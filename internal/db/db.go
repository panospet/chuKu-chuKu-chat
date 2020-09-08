package db

import (
	"fmt"

	"chuKu-chuKu-chat/internal/model"
)

type DbI interface {
	AddChannel(ch model.Channel) error
	GetChannel(name string) (model.Channel, error)
	ChannelLastMessages(name string, amount int) ([]model.Msg, error)
	GetChannels() ([]model.Channel, error)
	DeleteChannel(name string) error

	AddUser(user model.User) error
	GetUser(name string) (model.User, error)
	GetUsers() ([]model.User, error)
	RemoveUser(username string) error

	AddSubscription(username string, channelName string) error

	AddMessage(m model.Msg) error
}

type ErrChannelAlreadyExists struct {
	ChannelName string
}

func (e ErrChannelAlreadyExists) Error() string {
	return fmt.Sprintf("Channel with name '%s' already exists", e.ChannelName)
}
