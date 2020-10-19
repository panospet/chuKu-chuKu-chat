package common

import (
	"fmt"
	"sync"
)

var gray = []string{
	"#C0C0C0",
	"#A9A9A9",
	"#808080",
	"#696969",
	"#778899",
	"#708090",
	"#2F4F4F",
}

var mustard = []string{
	"#DEB887",
	"#D2B48C",
	"#BC8F8F",
	"#F4A460",
	"#DAA520",
	"#B8860B",
	"#CD853F",
	"#D2691E",
	"#8B4513",
	"#A0522D",
	"#A52A2A",
	"#800000",
}

var blue = []string{
	"#00FFFF",
	"#7FFFD4",
	"#40E0D0",
	"#48D1CC",
	"#00CED1",
	"#5F9EA0",
	"#4682B4",
	"#B0C4DE",
	"#B0E0E6",
	"#ADD8E6",
	"#87CEEB",
	"#87CEFA",
	"#00BFFF",
	"#1E90FF",
	"#6495ED",
	"#7B68EE",
	"#4169E1",
	"#0000FF",
	"#0000CD",
	"#00008B",
	"#000080",
	"#191970",
}

var red = []string{
	"#CD5C5C",
	"#F08080",
	"#FA8072",
	"#E9967A",
	"#FFA07A",
	"#DC143C",
	"#FF0000",
	"#B22222",
	"#8B0000",
}

var green = []string{
	"#ADFF2F",
	"#7FFF00",
	"#7CFC00",
	"#00FF00",
	"#32CD32",
	"#98FB98",
	"#90EE90",
	"#00FA9A",
	"#00FF7F",
	"#3CB371",
	"#2E8B57",
	"#228B22",
	"#008000",
	"#006400",
	"#9ACD32",
	"#6B8E23",
	"#808000",
	"#556B2F",
	"#66CDAA",
	"#8FBC8B",
	"#20B2AA",
	"#008B8B",
	"#008080",
}

var pink = []string{
	"#FFC0CB",
	"#FFB6C1",
	"#FF69B4",
	"#FF1493",
	"#C71585",
	"#DB7093",
}

var yellow = []string{
	"#BDB76B",
	"#FFD700",
}

var orange = []string{
	"#FFA07A",
	"#FF7F50",
	"#FF6347",
	"#FF4500",
	"#FF8C00",
	"#FFA500",
}

var violet = []string{
	"#E6E6FA",
	"#D8BFD8",
	"#DDA0DD",
	"#EE82EE",
	"#DA70D6",
	"#FF00FF",
	"#BA55D3",
	"#9370DB",
	"#663399",
	"#8A2BE2",
	"#9400D3",
	"#9932CC",
	"#8B008B",
	"#800080",
	"#4B0082",
	"#6A5ACD",
	"#483D8B",
	"#7B68EE",
}

type ColorPicker struct {
	Colors  []string
	Counter int
	mu      sync.Mutex
}

func NewColorPicker() *ColorPicker {
	var colors []string
	palette := [][]string{blue, red, green, mustard, gray, pink, yellow, orange, violet}
	counter := 0
	for {
		index := counter % len(palette)
		if len(palette[index]) > 0 {
			colors = append(colors, palette[index][0])
			palette[index] = palette[index][1:]
		} else {
			palette = append(palette[:index], palette[index+1:]...)
		}
		if len(palette) == 0 {
			break
		}
		counter++
	}

	return &ColorPicker{
		Colors:  colors,
		Counter: 0,
		mu:      sync.Mutex{},
	}
}

func (c *ColorPicker) ChooseNext() string {
	color := c.Colors[0]
	c.Colors = c.Colors[1:]
	if len(c.Colors) == 0 {
		c.Colors = GenerateColors()
	}
	return color
}

func GenerateColors() []string {
	fmt.Println("color palette has finished, regenerating!")
	var colors []string
	palette := [][]string{mustard, gray, blue, red, green, pink, yellow, orange, violet}
	counter := 0
	for {
		index := counter % len(palette)
		if len(palette[index]) > 0 {
			colors = append(colors, palette[index][0])
			palette[index] = palette[index][1:]
		} else {
			palette = append(palette[:index], palette[index+1:]...)
		}
		if len(palette) == 0 {
			break
		}
	}
	return colors
}
