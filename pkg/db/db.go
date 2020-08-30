package db

import (
	"chuKu-chuKu-chat/pkg/common"
	"chuKu-chuKu-chat/pkg/model"
	"errors"
)

type DbI interface {
	ChannelLastMessages(name string, amount int) ([]model.Msg, error)
	GetChannels() ([]model.Channel, error)
	CreateChannel(name string, description string) error
	DeleteChannel(name string) error

	GetUser(name string) (model.User, error)
	GetUsers() ([]model.User, error)
	AddUser(user model.User) error
}

type DummyDb struct {
	Channels map[string]model.Channel
	Users    map[string]model.User
}

func NewDummyDb() *DummyDb {
	g := model.Channel{
		Name:        "general",
		Description: "general discussion",
	}
	m := model.Channel{
		Name:        "metallica",
		Description: "metallica discussion",
	}
	return &DummyDb{
		Channels: map[string]model.Channel{"general": g, "metallica": m},
		Users:    map[string]model.User{},
	}
}

func (d *DummyDb) GetChannels() ([]model.Channel, error) {
	var out []model.Channel
	for _, c := range d.Channels {
		out = append(out, c)
	}
	return out, nil
}

func (d *DummyDb) ChannelLastMessages(name string, amount int) ([]model.Msg, error) {
	if _, ok := d.Channels[name]; !ok {
		return []model.Msg{}, errors.New("channel does not exist")
	}
	return common.GenerateRandomMessages(name, amount), nil
}

func (d *DummyDb) CreateChannel(name string, description string) error {
	c := model.Channel{
		Name:        name,
		Description: description,
	}
	d.Channels["name"] = c
	return nil
}

func (d *DummyDb) DeleteChannel(name string) error {
	if _, ok := d.Channels[name]; !ok {
		return errors.New("channel does not exist")
	}
	delete(d.Channels, name)
	return nil
}

func (d *DummyDb) GetUser(name string) (model.User, error) {
	user, ok := d.Users[name]
	if !ok {
		return model.User{}, errors.New("user does not exist")
	}
	return user, nil
}

func (d *DummyDb) GetUsers() ([]model.User, error) {
	var out []model.User
	for _, u := range d.Users {
		out = append(out, u)
	}
	return out, nil
}

func (d *DummyDb) AddUser(user model.User) error {
	if _, ok := d.Users[user.Username]; ok {
		return errors.New("user already exists")
	}
	d.Users[user.Username] = user
	return nil
}
