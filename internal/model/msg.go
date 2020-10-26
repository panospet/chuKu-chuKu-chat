package model

import "time"

type Msg struct {
	Id        string    `json:"id" db:"id"`
	Content   string    `json:"content" db:"content"`
	Channel   string    `json:"channel"`
	User      string    `json:"user"`
	UserColor string    `json:"user_color"`
	Timestamp time.Time `json:"timestamp" db:"sent_at"`

	Command int    `json:"command"`
	Err     string `json:"err,omitempty"`
}
