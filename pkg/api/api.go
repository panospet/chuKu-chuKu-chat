package api

import (
	"chuKu-chuKu-chat/pkg/user"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type App struct {
	upgrader       websocket.Upgrader
	redis          *redis.Client
}

func NewApi(redis *redis.Client) *App {
	return &App{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// todo fix cross origin
				return true
			},
		},
		redis: redis,
	}
}

func (a *App) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/channels", a.getChannels).Methods("GET")

	r.HandleFunc("/chat", a.ChatWebSocketHandler).Methods("GET")

	fmt.Println("serving")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func (a *App) getChannels(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, 200, []string{"general"})
}

func (a *App) ChatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("ERROR during websocket handshake:", err)
		return
	}

	username := r.URL.Query()["username"][0]
	u := user.NewUser(username)

	err = a.onConnect(r, conn, a.redis, *u)
	if err != nil {
		handleWSError(err, conn)
		return
	}

	closeCh := a.onDisconnect(conn, u)

	a.onChannelMessage(conn, u)

loop:
	for {
		select {
		case <-closeCh:
			break loop
		default:
			a.onUserCommand(conn, a.redis, u)
		}
	}
}

func (a *App) onConnect(r *http.Request, conn *websocket.Conn, rdb *redis.Client, u user.User) error {
	fmt.Println("connected from:", conn.RemoteAddr(), "user:", u.Username)
	err := u.Connect(rdb)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) onDisconnect(conn *websocket.Conn, u *user.User) chan struct{} {
	closeCh := make(chan struct{})

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("connection closed for user", u.Username)

		if err := u.Disconnect(); err != nil {
			return err
		}
		close(closeCh)
		return nil
	})

	return closeCh
}

func (a *App) onUserCommand(conn *websocket.Conn, rdb *redis.Client, u *user.User) error {
	var msg Msg

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

type Msg struct {
	Content string `json:"content,omitempty"`
	Channel string `json:"channel,omitempty"`
	Command int    `json:"command,omitempty"`
	Err     string `json:"err,omitempty"`
	User    string `json:"user,omitempty"`
}

func (a *App) onChannelMessage(conn *websocket.Conn, u *user.User) {
	go func() {
		for m := range u.MessageChan {
			fmt.Println("RECEIVED CHANNEEL MESSAGE", m)
			var t temp
			err := json.Unmarshal([]byte(m.Payload), &t)
			if err != nil {
				fmt.Println("ERROR unmashalling message:", err)
			}
			msg := Msg{
				Content: t.Content,
				Channel: t.Channel,
				User:    t.User,
			}
			if err := conn.WriteJSON(msg); err != nil {
				fmt.Println(err)
			}
		}

	}()
}

func handleWSError(err error, conn *websocket.Conn) {
	_ = conn.WriteJSON(Msg{Err: err.Error()})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
