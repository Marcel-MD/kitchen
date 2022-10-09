package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/Marcel-MD/kitchen/domain"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config()

	domain.SetConfig(cfg)
	orderChan := make(chan domain.Order, cfg.NrOfTables)
	orderList := domain.NewOrderList(orderChan)
	orderList.Run()

	r := mux.NewRouter()
	r.HandleFunc("/order", func(w http.ResponseWriter, r *http.Request) {
		var order domain.Order
		err := json.NewDecoder(r.Body).Decode(&order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		orderChan <- order

		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	http.ListenAndServe(":"+cfg.KitchenPort, r)
}

func config() domain.Config {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Logger = log.With().Caller().Logger()

	file, err := os.Open("config/cfg.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening menu.json")
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	var cfg domain.Config
	json.Unmarshal(byteValue, &cfg)

	return cfg
}
