package model

type PlayingInfo struct {
	ListenUrl    string `json:"listen_url"`
	IsLive       bool   `json:"is_live"`
	StreamerName string `json:"streamer_name"`
	SongArtist   string `json:"song_artist"`
	SongTitle    string `json:"song_title"`
	SongHistory  []Song `json:"song_history"`
}

type Song struct {
	SongArtist string `json:"song_artist"`
	SongTitle  string `json:"song_title"`
}
