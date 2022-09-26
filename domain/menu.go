package domain

import (
	"encoding/json"
	"io"
	"os"

	"github.com/rs/zerolog/log"
)

type Menu struct {
	FoodsCount int
	Foods      []Food
}

type Food struct {
	Id               int    `json:"id"`
	Name             string `json:"name"`
	PreparationTime  int    `json:"preparation_time"`
	Complexity       int    `json:"complexity"`
	CookingApparatus string `json:"cooking_apparatus"`
}

func GetMenu() Menu {
	file, err := os.Open("config/menu.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening menu.json")
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	var menu Menu
	json.Unmarshal(byteValue, &menu)

	menu.FoodsCount = len(menu.Foods)
	return menu
}
