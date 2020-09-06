package model

import (
	"errors"
	"time"
)

type Channel struct {
	Id          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Creator     string    `json:"creator" db:"creator"`
	IsPrivate   bool      `json:"is_private" db:"is_private"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
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
