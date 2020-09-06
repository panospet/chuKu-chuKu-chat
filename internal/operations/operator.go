package operations

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/internal/common"
	"chuKu-chuKu-chat/internal/model"
)

type Operator interface {
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

type DummyOperator struct {
	Channels map[string]model.Channel
	Users    map[string]model.User
	rdb      *redis.Client
}

func NewDummyOperator(rdb *redis.Client) (*DummyOperator, error) {
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
	u.AddChannel("metallica")
	u.AddChannel("general")
	err = u.RefreshChannels(rdb)
	if err != nil {
		return nil, nil
	}
	return &DummyOperator{
		Channels: map[string]model.Channel{"general": g, "metallica": m},
		Users:    map[string]model.User{u.Username: *u},
		rdb:      rdb,
	}, nil
}

func (d *DummyOperator) GetChannel(name string) (model.Channel, error) {
	c, ok := d.Channels[name]
	if !ok {
		return model.Channel{}, errors.New("channel does not exist")
	}
	return c, nil
}

func (d *DummyOperator) GetChannels() ([]model.Channel, error) {
	var out []model.Channel
	for _, c := range d.Channels {
		out = append(out, c)
	}
	return out, nil
}

func (d *DummyOperator) ChannelLastMessages(name string, amount int) ([]model.Msg, error) {
	if _, err := d.GetChannel(name); err != nil {
		return []model.Msg{}, err
	}
	return common.GenerateRandomMessages(name, amount), nil
}

type ErrChannelAlreadyExists struct {
	ChannelName string
}

func (e ErrChannelAlreadyExists) Error() string {
	return fmt.Sprintf("Channel with name '%s' already exists", e.ChannelName)
}

func (d *DummyOperator) CreateChannel(name string, description string, creator string) error {
	user, ok := d.Users[creator]
	if !ok {
		return errors.New("user does not exist")
	}
	c := model.Channel{
		Name:        name,
		Description: description,
		Creator:     creator,
	}
	if _, ok := d.Channels[name]; ok {
		return ErrChannelAlreadyExists{ChannelName: name}
	}
	d.Channels[name] = c
	user.AddChannel(c.Name)
	return user.RefreshChannels(d.rdb)
}

func (d *DummyOperator) DeleteChannel(name string) error {
	if _, err := d.GetChannel(name); err != nil {
		return err
	}
	delete(d.Channels, name)
	return nil
}

func (d *DummyOperator) GetUser(name string) (model.User, error) {
	user, ok := d.Users[name]
	if !ok {
		return model.User{}, errors.New("user does not exist")
	}
	return user, nil
}

func (d *DummyOperator) GetUsers() ([]model.User, error) {
	var out []model.User
	for _, u := range d.Users {
		out = append(out, u)
	}
	return out, nil
}

func (d *DummyOperator) AddUser(user model.User) error {
	if _, ok := d.Users[user.Username]; ok {
		return errors.New("user already exists")
	}
	d.Users[user.Username] = user
	return nil
}

func (d *DummyOperator) RemoveUser(username string) error {
	if _, ok := d.Users[username]; !ok {
		return errors.New("user does not exist")
	}
	delete(d.Users, username)
	return nil
}

func (d *DummyOperator) Subscription(username string, channelName string) error {
	u, ok := d.Users[username]
	if !ok {
		return errors.New("user does not exist")
	}
	_, ok = d.Channels[channelName]
	if !ok {
		return errors.New("channel does not exist")
	}
	u.AddChannel(channelName)
	err := u.RefreshChannels(d.rdb)
	if err != nil {
		return err
	}
	return nil
}
