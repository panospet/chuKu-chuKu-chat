package common

import (
	"math/rand"
	"time"

	"github.com/tjarratt/babble"

	"chuKu-chuKu-chat/internal/model"
)

func GenerateRandomMessages(channelName string, amount int) []model.Msg {
	babbler := babble.NewBabbler()
	randomUsers := []string{"panos", "odi", "steve", "pitsikokos", "mario"}
	babbler.Separator = " "
	rand.Seed(time.Now().UnixNano())
	var output []model.Msg
	for i := 0; i < amount; i++ {
		babbler.Count = rand.Intn(6) + 1
		output = append(output, model.Msg{
			Content: babbler.Babble(),
			Channel: channelName,
			Command: 0,
			User:    randomUsers[rand.Intn(5)],
			Timestamp: time.Now(),
		})
	}
	return output
}
