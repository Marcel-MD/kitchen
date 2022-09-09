package main

import (
	"os"

	"github.com/Marcel-MD/kitchen/domain"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config()

	r := gin.Default()
	r.POST("/order", func(c *gin.Context) {
		var order domain.Order

		if err := c.ShouldBindJSON(&order); err != nil {
			log.Err(err).Msg("Error binding JSON")
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "Order received"})
	})
	r.Run(":8081")
}

func config() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading .env file")
	}
}
