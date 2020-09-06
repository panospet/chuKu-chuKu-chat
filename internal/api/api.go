package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/go-redis/redis/v7"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"chuKu-chuKu-chat/internal/operations"
	"chuKu-chuKu-chat/internal/model"
)

type App struct {
	upgrader websocket.Upgrader
	rdb      *redis.Client
	operator operations.Operator
}

func NewApp(redis *redis.Client, db operations.Operator) *App {
	return &App{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// todo fix cross origin
				return true
			},
		},
		rdb:      redis,
		operator: db,
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

	r.HandleFunc("/chat", a.chatWebSocketHandler).Methods("GET")

	// todo configure allowed origins
	//originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	fmt.Println("serving!")
	log.Fatal(http.ListenAndServe(":8000", handlers.CORS(originsOk, methodsOk)(r)))
}

func (a *App) getChannels(w http.ResponseWriter, r *http.Request) {
	channels, err := a.operator.GetChannels()
	if err != nil {
		respondWithError(w, 404, "channel does not exist")
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
	messages, err := a.operator.ChannelLastMessages(channelName, amount)
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
	if err := c.Validate(); err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	err = a.operator.AddChannel(c.Name, c.Description, c.Creator)
	if err != nil {
		if er, ok := err.(operations.ErrChannelAlreadyExists); ok {
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
	err := a.operator.DeleteChannel(channelName)
	if err != nil {
		respondWithError(w, 500, "channel deletion failure")
		return
	}
}

func (a *App) getUserChannels(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]
	user, err := a.operator.GetUser(username)
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
	users, err := a.operator.GetUsers()
	if err != nil {
		respondWithError(w, 500, "an error occured")
		return
	}
	var out []UserJson
	for _, u := range users {
		out = append(out, UserJson{Name: u.Username})
	}
	respondWithJSON(w, 200, out)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["user"]
	u, err := a.operator.GetUser(username)
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
	err = a.operator.AddSubscription(s.User, s.Channel)
	if err != nil {
		respondWithError(w, 500, "an error occured")
		return
	}
	respondWithJSON(w, 201, SuccessMessage{Message: fmt.Sprintf("User %s was subscribed to channel %s"+
		" successfully", s.User, s.Channel)})
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
