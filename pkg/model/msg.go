package model

type Msg struct {
	Content   string `json:"content"`
	Channel   string `json:"channel"`
	Command   int    `json:"command"`
	Err       string `json:"err,omitempty"`
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
}
