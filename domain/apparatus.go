package domain

import (
	"encoding/json"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog/log"
)

type Apparatus struct {
	Name     string `json:"name"`
	Quantity int64  `json:"quantity"`
	used     int64
	mu       sync.Mutex
}

func (a *Apparatus) Use() bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.used < a.Quantity {
		a.used++
		return true
	}

	return false
}

func (a *Apparatus) Release() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.used--
}

type apparatuses struct {
	Apparatuses []Apparatus
}

func GetApparatusesMap() map[string]*Apparatus {
	file, err := os.Open("config/apparatuses.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening apparatuses.json")
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)
	var a apparatuses
	json.Unmarshal(byteValue, &a)

	apparatusesMap := make(map[string]*Apparatus)
	for i := range a.Apparatuses {
		apparatusesMap[a.Apparatuses[i].Name] = &a.Apparatuses[i]
	}

	return apparatusesMap
}
