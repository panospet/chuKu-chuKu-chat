package common

import (
	"github.com/google/uuid"
	"math/rand"
	"time"

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
			Command:   0,
			User:      users[rand.Intn(len(users))],
			Timestamp: time.Now(),
		})
	}
	return output
}
