package db

import (
	"errors"
	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/pkg/common"
	"chuKu-chuKu-chat/pkg/model"
)

type OperationsI interface {
	ChannelLastMessages(name string, amount int) ([]model.Msg, error)
	GetChannels() ([]model.Channel, error)
	CreateChannel(name string, description string, creator string) error
	DeleteChannel(name string) error
	GetChannel(name string) (model.Channel, error)

	GetUser(name string) (model.User, error)
	GetUsers() ([]model.User, error)
	AddUser(user model.User) error
	RemoveUser(username string) error

	Subscription(username string, channelName string) error
}

type DummyOperations struct {
	Channels map[string]model.Channel
	Users    map[string]model.User
	rdb      *redis.Client
}

func NewDummyOperations(rdb *redis.Client) (*DummyOperations, error) {
	g := model.Channel{
		Name:        "general",
		Description: "general discussion",
		Creator:     "admin",
	}
	var err error
	m := model.Channel{
		Name:        "metallica",
		Description: "metallica discussion",
		Creator:     "admin",
	}
	u := model.NewUser("admin")
	err = u.SubscribeToChannel("metallica", rdb)
	if err != nil {
		return nil, nil
	}
	err = u.SubscribeToChannel("general", rdb)
	if err != nil {
		return nil, nil
	}
	return &DummyOperations{
		Channels: map[string]model.Channel{"general": g, "metallica": m},
		Users:    map[string]model.User{u.Username: *u},
		rdb:      rdb,
	}, nil
}

func (d *DummyOperations) GetChannel(name string) (model.Channel, error) {
	c, ok := d.Channels[name]
	if !ok {
		return model.Channel{}, errors.New("channel does not exist")
	}
	return c, nil
}

func (d *DummyOperations) GetChannels() ([]model.Channel, error) {
	var out []model.Channel
	for _, c := range d.Channels {
		out = append(out, c)
	}
	return out, nil
}

func (d *DummyOperations) ChannelLastMessages(name string, amount int) ([]model.Msg, error) {
	if _, err := d.GetChannel(name); err != nil {
		return []model.Msg{}, err
	}
	return common.GenerateRandomMessages(name, amount), nil
}

func (d *DummyOperations) CreateChannel(name string, description string, creator string) error {
	user, ok := d.Users[creator]
	if !ok {
		return errors.New("user does not exist")
	}
	c := model.Channel{
		Name:        name,
		Description: description,
		Creator:     creator,
	}
	d.Channels["name"] = c
	return user.SubscribeToChannel(name, d.rdb)
}

func (d *DummyOperations) DeleteChannel(name string) error {
	if _, err := d.GetChannel(name); err != nil {
		return err
	}
	delete(d.Channels, name)
	return nil
}

func (d *DummyOperations) GetUser(name string) (model.User, error) {
	user, ok := d.Users[name]
	if !ok {
		return model.User{}, errors.New("user does not exist")
	}
	return user, nil
}

func (d *DummyOperations) GetUsers() ([]model.User, error) {
	var out []model.User
	for _, u := range d.Users {
		out = append(out, u)
	}
	return out, nil
}

func (d *DummyOperations) AddUser(user model.User) error {
	if _, ok := d.Users[user.Username]; ok {
		return errors.New("user already exists")
	}
	d.Users[user.Username] = user
	return nil
}

func (d *DummyOperations) RemoveUser(username string) error {
	if _, ok := d.Users[username]; !ok {
		return errors.New("user does not exist")
	}
	delete(d.Users, username)
	return nil
}

func (d *DummyOperations) Subscription(username string, channelName string) error {
	u, ok := d.Users[username]
	if !ok {
		return errors.New("user does not exist")
	}
	_, ok = d.Channels[channelName]
	if !ok {
		return errors.New("channel does not exist")
	}
	err := u.SubscribeToChannel(channelName, d.rdb)
	if err != nil {
		return err
	}
	return nil
}
