package api

import (
	"chuKu-chuKu-chat/pkg/db"
	"chuKu-chuKu-chat/pkg/model"
	"chuKu-chuKu-chat/pkg/user"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type App struct {
	upgrader websocket.Upgrader
	redis    *redis.Client
	db       db.DbI
}

func NewApi(redis *redis.Client, db db.DbI) *App {
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
		db:    db,
	}
}

func (a *App) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/health", a.healthCheck).Methods("GET")

	r.HandleFunc("/channels", a.getChannels).Methods("GET")
	r.HandleFunc("/channels", a.createChannel).Methods("POST")
	r.HandleFunc("/channels/{channelName}", a.deleteChannel).Methods("DELETE")
	r.HandleFunc("/channels/{channelName}/lastMessages", a.getChannelLastMessages).Methods("GET")

	r.HandleFunc("/users/{user}/channels", a.getUserChannels).Methods("GET")

	r.HandleFunc("/chat", a.chatWebSocketHandler).Methods("GET")

	fmt.Println("serving!")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func (a *App) getChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := a.db.GetChannels()
	if err != nil {
		respondWithError(w, 404, "channel does not exist")
		return
	}
	respondWithJSON(w, 200, channels)
}

func (a *App) chatWebSocketHandler(w http.ResponseWriter, r *http.Request) {
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
	var msg model.Msg

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

func (a *App) onChannelMessage(conn *websocket.Conn, u *user.User) {
	go func() {
		for m := range u.MessageChan {
			fmt.Println("RECEIVED CHANNEEL MESSAGE", m)
			var t temp
			err := json.Unmarshal([]byte(m.Payload), &t)
			if err != nil {
				fmt.Println("ERROR unmashalling message:", err)
			}
			msg := model.Msg{
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

func (a *App) getChannelLastMessages(w http.ResponseWriter, r *http.Request) {
	channelName := mux.Vars(r)["channelName"]
	amount, err := strconv.Atoi(r.URL.Query().Get("amount"))
	if err != nil {
		respondWithError(w, 400, "bad amount given")
		return
	}
	messages, err := a.db.ChannelLastMessages(channelName, amount)
	if err != nil {
		// todo log error
		respondWithError(w, 500, "an error occured")
		return
	}
	respondWithJSON(w, 200, messages)
}

type SuccessMessage struct {
	Message string `json:"message"`
}

func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, 200, SuccessMessage{Message: "Hi, I'm fine, and you?"})
}

func (a *App) createChannel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 400, "failed to read request body")
		return
	}
	var c model.Channel
	err = json.Unmarshal(b, &c)
	if err != nil {
		respondWithError(w, 400, "failed to read request body")
		return
	}
	err = a.db.CreateChannel(c.Name, c.Description)
	if err != nil {
		respondWithError(w, 500, "channel creation failure")
		return
	}
	respondWithJSON(w, 201, SuccessMessage{Message: fmt.Sprintf("channel with name %s "+
		"has been created successfully", c.Name)})
}

func (a *App) deleteChannel(w http.ResponseWriter, r *http.Request) {
	channelName := mux.Vars(r)["channelName"]
	if channelName == "" {
		respondWithError(w, 400, "please give a valid channel name")
		return
	}
	err := a.db.DeleteChannel(channelName)
	if err != nil {
		respondWithError(w, 500, "channel deletion failure")
		return
	}
}

func (a *App) getUserChannels(w http.ResponseWriter, r *http.Request) {

}

func handleWSError(err error, conn *websocket.Conn) {
	_ = conn.WriteJSON(model.Msg{Err: err.Error()})
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
