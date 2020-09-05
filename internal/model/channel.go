package model

import (
	"errors"
	"time"
)

type Channel struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Creator     string    `json:"creator"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
}

func (c *Channel) Validate() error {
	if c.Name == "" {
		return errors.New("channel name cannot be blank")
	}
	if c.Creator == "" {
		return errors.New("channel creator cannot be blank")
	}
	return nil
}
