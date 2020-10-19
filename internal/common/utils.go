package common

import (
	"math/rand"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/tjarratt/babble"

	"chuKu-chuKu-chat/internal/model"
)

func GenerateRandomMessages(channelName string, amount int, users ...string) []model.Msg {
	babbler := babble.NewBabbler()
	babbler.Separator = " "
	rand.Seed(time.Now().UnixNano())
	var output []model.Msg
	for i := 0; i < amount; i++ {
		babbler.Count = rand.Intn(6) + 1
		output = append(output, model.Msg{
			Id:        uuid.New().String(),
			Content:   babbler.Babble(),
			Channel:   channelName,
			User:      users[rand.Intn(len(users))],
			Timestamp: time.Now().Add(time.Duration(-rand.Intn(48)) * time.Hour),
		})
	}
	sort.Slice(output[:], func(i, j int) bool {
		return output[i].Timestamp.UnixNano() < output[j].Timestamp.UnixNano()
	})
	return output
}

func Shuffle(a []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
}
