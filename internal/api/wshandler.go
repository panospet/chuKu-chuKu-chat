package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"chuKu-chuKu-chat/internal/model"
)

func (a *App) chatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR during websocket handshake:", err)
		return
	}
	username := r.URL.Query()["username"][0]
	var u model.User
	newUser := model.NewUser(username)
	err = a.db.AddUser(*newUser)
	if err != nil {
		log.Println("error adding user", err)
		handleWSError(errors.New("user already exists"), conn)
		return
	}
	u = *newUser

	err = a.onConnect(r, conn, a.rdb, *newUser)
	if err != nil {
		handleWSError(err, conn)
		return
	}

	closeCh := a.onDisconnect(conn, &u)

	a.onChannelMessage(conn, &u)

loop:
	for {
		select {
		case <-closeCh:
			break loop
		default:
			err := a.onUserCommand(conn, a.rdb)
			if err != nil {
				fmt.Println("on user command error", err)
				break loop
			}
		}
	}
}

func (a *App) onConnect(r *http.Request, conn *websocket.Conn, rdb *redis.Client, u model.User) error {
	fmt.Println("connected from:", conn.RemoteAddr(), "user:", u.Username)
	err := u.ConnectToPubSub(rdb)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) onDisconnect(conn *websocket.Conn, u *model.User) chan struct{} {
	closeCh := make(chan struct{})

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("closing connection for user", u.Username)

		if err := u.Disconnect(); err != nil {
			return err
		}
		close(closeCh)

		return nil
	})
	fmt.Println("connection closed for user", u.Username)
	return closeCh
}

func (a *App) onUserCommand(conn *websocket.Conn, rdb *redis.Client) error {
	var msg model.Msg
	msg.Timestamp = time.Now()
	if err := conn.ReadJSON(&msg); err != nil {
		handleWSError(err, conn)
		return err
	}
	msg.Id = uuid.New().String()
	msgB, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("ERROR CHAT:", err)
		return err
	}
	fmt.Println("Chat function with msg:", string(msgB))

	go func() {
		err := a.db.AddMessage(msg)
		if err != nil {
			fmt.Println("ERROR: could not store message")
		}
	}()

	return rdb.Publish(msg.Channel, string(msgB)).Err()
}

type temp struct {
	Content   string    `json:"content"`
	Channel   string    `json:"channel"`
	User      string    `json:"user"`
	Command   int       `json:"command"`
	Timestamp time.Time `json:"timestamp"`
}

func (a *App) onChannelMessage(conn *websocket.Conn, u *model.User) {
	go func() {
		for m := range u.MessageChan {
			fmt.Println("RECEIVED CHANNEEL MESSAGE", m)
			var t temp
			err := json.Unmarshal([]byte(m.Payload), &t)
			if err != nil {
				fmt.Println("ERROR unmashalling message:", err)
			}
			msg := model.Msg{
				Content:   t.Content,
				Channel:   t.Channel,
				User:      t.User,
				Timestamp: t.Timestamp,
			}
			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println(err)
			}
		}

	}()
}
