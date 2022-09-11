package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Marcel-MD/kitchen/domain"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config()

	domain.SetConfig(cfg)
	orderChan := make(chan domain.Order, cfg.NrOfTables)
	menu := domain.GetMenu()
	orderList := domain.NewOrderList(orderChan, menu)
	orderList.Run()

	r := gin.Default()
	r.POST("/order", func(c *gin.Context) {
		var order domain.Order

		if err := c.ShouldBindJSON(&order); err != nil {
			log.Err(err).Msg("Error binding JSON")
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		orderChan <- order

		c.JSON(200, gin.H{"message": "Order received"})
	})
	r.Run(":8081")
}

func config() domain.Config {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()

	file, err := os.Open("config/cfg.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening menu.json")
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)
	var cfg domain.Config
	json.Unmarshal(byteValue, &cfg)

	return cfg
}
