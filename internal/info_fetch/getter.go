package info_fetch

import (
	"chuKu-chuKu-chat/internal/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Getter interface {
	Get() (model.PlayingInfo, error)
}

type AzuraGetter struct {
	url    string
	client http.Client
}

func NewAzuraGetter(url string) *AzuraGetter {
	c := http.Client{
		Timeout: 15 * time.Second,
	}
	return &AzuraGetter{url: url, client: c}
}

type AzuraLive struct {
	IsLive       bool   `json:"is_live"`
	StreamerName string `json:"streamer_name"`
}

type AzuraNowPlayingResponse struct {
	Station     AzuraStation           `json:"station"`
	Listeners   map[string]interface{} `json:"listeners"`
	Live        AzuraLive              `json:"live"`
	NowPlaying  AzuraSongWrap          `json:"now_playing"`
	SongHistory []AzuraSongWrap        `json:"song_history"`
}

type AzuraStation struct {
	Id        int    `json:"id"`
	Shortcode string `json:"shortcode"`
	ListenUrl string `json:"listen_url"`
}

type AzuraSongWrap struct {
	Streamer string    `json:"streamer"`
	Song     AzuraSong `json:"song"`
}

type AzuraSong struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

func (a *AzuraGetter) Get() (model.PlayingInfo, error) {
	req, err := http.NewRequest("GET", a.url, nil)
	if err != nil {
		return model.PlayingInfo{}, err
	}
	res, err := a.client.Do(req)
	if err != nil {
		return model.PlayingInfo{}, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return model.PlayingInfo{}, errors.New(fmt.Sprintf("bad status code: %d", res.StatusCode))
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return model.PlayingInfo{}, err
	}
	res.Body.Close()
	var response []AzuraNowPlayingResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return model.PlayingInfo{}, err
	}
	var out model.PlayingInfo
	for _, a := range response {
		if a.Station.Shortcode == "kickit-radio" {
			out.IsLive = a.Live.IsLive
			out.StreamerName = a.Live.StreamerName
			out.ListenUrl = a.Station.ListenUrl
			out.SongArtist = a.NowPlaying.Song.Artist
			out.SongTitle = a.NowPlaying.Song.Title
			for _, s := range a.SongHistory {
				out.SongHistory = append(out.SongHistory, model.Song{
					SongArtist: s.Song.Artist,
					SongTitle:  s.Song.Title,
				})
			}
		}
	}
	return out, nil
}
