package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/websocket"

	"chuKu-chuKu-chat/pkg/model"
)

func (a *App) chatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR during websocket handshake:", err)
		return
	}
	username := r.URL.Query()["username"][0]
	var u model.User
	u, err = a.db.GetUser(username)
	if err != nil {
		newUser := model.NewUser(username)
		err = a.db.AddUser(*newUser)
		if err != nil {
			handleWSError(err, conn)
			return
		}
		u = *newUser
	}

	err = a.onConnect(r, conn, a.rdb, u)
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

		return a.db.RemoveUser(u.Username)
		//return nil
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
	msgB, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("ERROR CHAT:", err)
		return err
	}
	fmt.Println("Chat function with msg:", string(msgB))

	return rdb.Publish(msg.Channel, string(msgB)).Err()
}

type temp struct {
	Content string
	Channel string
	User    string
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
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println(err)
			}
		}

	}()
}
