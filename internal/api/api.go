package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"chuku-chuku-chat/config"
	"chuku-chuku-chat/internal/common"
	"chuku-chuku-chat/internal/db"
	"chuku-chuku-chat/internal/info_fetch"
	"chuku-chuku-chat/internal/model"
)

type App struct {
	upgrader    websocket.Upgrader
	rdb         *redis.Client
	db          db.DbI
	infoGetter  info_fetch.Getter
	mode        string
	colorPicker *common.ColorPicker
}

func NewApp(mode string, redis *redis.Client, db db.DbI, infoGetter info_fetch.Getter) *App {
	return &App{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// todo fix cross origin
				return true
			},
		},
		rdb:         redis,
		db:          db,
		infoGetter:  infoGetter,
		mode:        mode,
		colorPicker: common.NewColorPicker(),
	}
}

func (a *App) Run() {
	r := mux.NewRouter()

	r.HandleFunc("/health", a.healthCheck).Methods("GET")

	r.HandleFunc("/channels", a.getChannels).Methods("GET")
	r.HandleFunc("/channels", a.createChannel).Methods("POST")
	r.HandleFunc("/channels/subscription", a.subscription).Methods("POST")
	r.HandleFunc("/channels/{channelName}", a.deleteChannel).Methods("DELETE")
	r.HandleFunc("/channels/{channelName}/lastMessages", a.getChannelLastMessages).Methods("GET")

	r.HandleFunc("/users", a.getUsers).Methods("GET")
	r.HandleFunc("/users/{user}", a.getUser).Methods("GET")
	r.HandleFunc("/users/{user}/channels", a.getUserChannels).Methods("GET")
	r.HandleFunc("/users/{user}", a.deleteUser).Methods("DELETE")

	r.HandleFunc("/chat", a.chatWebSocketHandler).Methods("GET")

	r.HandleFunc("/info", a.getNowPlayingInfo).Methods("GET")

	// todo configure allowed origins
	//originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE"})

	go a.Monitor()
	go a.BroadcastNowPlayingInfo()
	fmt.Println("serving!")
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, methodsOk)(r)))
}

func (a *App) BroadcastMessage(p Payload, c string) {
	b, _ := json.Marshal(p)
	a.rdb.Publish(c, string(b))
}

func (a *App) Monitor() {
	var amount int
	if a.mode == config.DummyMode {
		amount = 1500
	} else {
		amount = 48
	}
	if err := a.db.ClearOldMessages(amount); err != nil {
		log.Println("Monitor: error while clearing old messages:", err)
	} else {
		log.Println("Monitor: cleared old messages with success", time.Now().String())
	}
	for now := range time.Tick(12 * time.Hour) {
		if err := a.db.ClearOldMessages(amount); err != nil {
			log.Println("Monitor: error while clearing old messages:", err)
		} else {
			log.Println("Monitor: cleared old messages with success", now.String())
		}
	}
}

func (a *App) BroadcastNowPlayingInfo() {
	info, err := a.infoGetter.Get()
	if err != nil {
		log.Println("cannot get now playing info!", err)
		return
	}
	artist := info.SongArtist
	title := info.SongTitle
	nowPlaying := fmt.Sprintf("Now playing: %s - %s", artist, title)
	a.BroadcastMessage(Payload{
		Content:   nowPlaying,
		Channel:   "general",
		User:      db.KickItBotUsername,
		UserColor: "#000000",
		Command:   2,
		Timestamp: time.Now(),
	}, "general")
	for now := range time.Tick(15 * time.Second) {
		info, err := a.infoGetter.Get()
		if err != nil {
			log.Println("cannot get now playing info!", err)
		}
		if info.SongArtist != artist || info.SongTitle != title {
			artist = info.SongArtist
			title = info.SongTitle
			nowPlaying := fmt.Sprintf("Now playing: %s - %s", artist, title)
			a.BroadcastMessage(Payload{
				Content:   nowPlaying,
				Channel:   "general",
				User:      db.KickItBotUsername,
				UserColor: "#000000",
				Command:   2,
				Timestamp: now,
			}, "general")
		}
	}
}

func (a *App) getChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := a.db.GetChannels()
	if err != nil {
		respondWithError(w, 404, "no channels found")
		return
	}
	respondWithJSON(w, 200, channels)
}

func (a *App) getChannelLastMessages(w http.ResponseWriter, r *http.Request) {
	channelName := mux.Vars(r)["channelName"]
	amount, err := strconv.Atoi(r.URL.Query().Get("amount"))
	if err != nil {
		respondWithError(w, 400, "bad amount given")
		return
	}
	messages, err := a.db.ChannelLastMessages(channelName, amount)
	if err != nil && err != sql.ErrNoRows {
		log.Println("error getting channel last messages:", err)
		respondWithError(w, 500, "an error occured")
		return
	} else if err == sql.ErrNoRows {
		respondWithJSON(w, 200, []model.Msg{})
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
	if err := c.Validate(); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	err = a.db.AddChannel(c)
	if err != nil {
		if er, ok := err.(db.ErrChannelAlreadyExists); ok {
			fmt.Println("CONFLICTAKI")
			respondWithError(w, http.StatusConflict, er.Error())
			return
		}
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
		log.Println("error deleting channel", err)
		respondWithError(w, 500, "channel deletion failure")
		return
	}
}

func (a *App) getUserChannels(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]
	user, err := a.db.GetUser(username)
	if err != nil {
		respondWithError(w, 400, "user not found")
		return
	}
	respondWithJSON(w, 200, user.GetChannels())
}

type UserJson struct {
	Name string `json:"name"`
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	users, err := a.db.GetUsers()
	if err != nil {
		log.Println("error getting users:", err)
		respondWithError(w, 500, "an error occured")
		return
	}
	respondWithJSON(w, 200, users)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]
	u, err := a.db.GetUser(username)
	if err != nil {
		respondWithError(w, 404, "user not found")
	}
	respondWithJSON(w, 200, u)
}

type Subscription struct {
	User    string
	Channel string
}

func (a *App) subscription(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, 400, "failed to read request body")
		return
	}
	var s Subscription
	err = json.Unmarshal(b, &s)
	if err != nil {
		respondWithError(w, 400, "failed to read request body")
		return
	}
	err = a.db.AddSubscription(s.User, s.Channel)
	if err != nil {
		log.Println("error adding subscription:", err)
		respondWithError(w, 500, "an error occured")
		return
	}
	respondWithJSON(w, 201, SuccessMessage{Message: fmt.Sprintf("User %s was subscribed to channel %s"+
		" successfully", s.User, s.Channel)})
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]
	if err := a.db.RemoveUser(username); err != nil {
		respondWithError(w, 404, "user not found")
	}
	respondWithJSON(w, 200, SuccessMessage{Message: "user was deleted successfully"})
}

func (a *App) getNowPlayingInfo(w http.ResponseWriter, r *http.Request) {
	info, err := a.infoGetter.Get()
	if err != nil {
		fmt.Println("error getting info", err)
		respondWithError(w, 500, "error getting info")
	}
	respondWithJSON(w, 200, info)
}

func handleWSError(err error, conn *websocket.Conn) {
	_ = conn.WriteJSON(model.Msg{Err: err.Error()})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return errors.New(fmt.Sprintf("error unmashaling: %s", err))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}
