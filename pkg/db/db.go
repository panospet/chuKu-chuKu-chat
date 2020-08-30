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
}

type DummyDb struct {
	Channels map[string]model.Channel
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
	return &DummyDb{Channels: map[string]model.Channel{"general": g, "metallica": m}}
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
