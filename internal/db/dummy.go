package db

import (
	"errors"
	"github.com/google/uuid"
	"time"

	"github.com/go-redis/redis/v7"

	"chuKu-chuKu-chat/internal/model"
)

const KickItBotUsername = "KickIt-bot"

type DummyDb struct {
	Channels map[string]model.Channel
	Users    map[string]model.User
	Messages []model.Msg
	rdb      *redis.Client
}

func NewDummyDb(rdb *redis.Client) (*DummyDb, error) {
	g := model.Channel{
		Name:        "general",
		Description: "general discussion",
		Creator:     KickItBotUsername,
	}
	u := model.NewUser(KickItBotUsername, "#000000", "general")
	err := u.RefreshChannels(rdb)
	if err != nil {
		return nil, errors.New("error refreshing channels in startup: " + err.Error())
	}
	msgs := []model.Msg{{
		Id:        uuid.New().String(),
		Content:   "Chatting started!",
		Channel:   "general",
		User:      KickItBotUsername,
		UserColor: u.ColorCode,
		Timestamp: time.Now(),
	}}
	return &DummyDb{
		Channels: map[string]model.Channel{"general": g},
		Users:    map[string]model.User{u.Username: *u},
		rdb:      rdb,
		Messages: msgs,
	}, nil
}

func (d *DummyDb) AddMessage(msg model.Msg) error {
	d.Messages = append(d.Messages, msg)
	return nil
}

func (d *DummyDb) GetChannel(name string) (model.Channel, error) {
	c, ok := d.Channels[name]
	if !ok {
		return model.Channel{}, errors.New("channel does not exist")
	}
	return c, nil
}

func (d *DummyDb) GetChannels() ([]model.Channel, error) {
	var out []model.Channel
	for _, c := range d.Channels {
		out = append(out, c)
	}
	return out, nil
}

func (d *DummyDb) ChannelLastMessages(channelName string, amount int) ([]model.Msg, error) {
	if _, err := d.GetChannel(channelName); err != nil {
		return []model.Msg{}, err
	}
	if amount >= len(d.Messages) {
		return d.Messages, nil
	}
	return d.Messages[len(d.Messages)-amount-1:], nil
}

func (d *DummyDb) AddChannel(ch model.Channel) error {
	user, ok := d.Users[ch.Creator]
	if !ok {
		return errors.New("user does not exist")
	}
	if _, ok := d.Channels[ch.Name]; ok {
		return ErrChannelAlreadyExists{ChannelName: ch.Name}
	}
	d.Channels[ch.Name] = ch
	user.AddChannel(ch.Name)
	return user.RefreshChannels(d.rdb)
}

func (d *DummyDb) DeleteChannel(name string) error {
	if _, err := d.GetChannel(name); err != nil {
		return err
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

func (d *DummyDb) UpdateUserLastMessage(username string) {
	if u, ok := d.Users[username]; ok {
		u.LastMessageAt = time.Now()
		d.Users[username] = u
	}
}

func (d *DummyDb) RemoveUser(username string) error {
	if _, ok := d.Users[username]; !ok {
		return errors.New("user does not exist")
	}
	delete(d.Users, username)
	return nil
}

func (d *DummyDb) AddSubscription(username string, channelName string) error {
	u, ok := d.Users[username]
	if !ok {
		return errors.New("user does not exist")
	}
	_, ok = d.Channels[channelName]
	if !ok {
		return errors.New("channel does not exist")
	}
	u.AddChannel(channelName)
	return u.RefreshChannels(d.rdb)
}

func (d *DummyDb) ClearOldMessages(amount int) error {
	if len(d.Messages) > amount {
		d.Messages = d.Messages[len(d.Messages)-500:]
	}
	return nil
}
