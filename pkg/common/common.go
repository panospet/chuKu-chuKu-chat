package common

import (
	"chuKu-chuKu-chat/pkg/model"
	"github.com/tjarratt/babble"
	"math/rand"
)

var babbler = babble.NewBabbler()
var randomUsers = []string{"panos", "odi", "poutsikletos", "pitsikokos", "tsoutsounakis"}

func init() {
	babbler.Separator = " "
}

func GenerateRandomMessages(channelName string, amount int) []model.Msg {
	rand.Seed(2384)
	var output []model.Msg
	for i := 0; i < amount; i++ {
		babbler.Count = rand.Intn(6) + 1
		output = append(output, model.Msg{
			Content: babbler.Babble(),
			Channel: channelName,
			Command: 0,
			User:    randomUsers[rand.Intn(5)],
		})
	}
	return output
}
