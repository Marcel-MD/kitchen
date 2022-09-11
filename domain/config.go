package domain

type Config struct {
	TimeUnit   int `json:"time_unit"`
	NrOfTables int `json:"nr_of_tables"`

	DiningHallUrl string `json:"dining_hall_url"`
}

var cfg Config = Config{
	TimeUnit:   1000,
	NrOfTables: 10,

	DiningHallUrl: "http://dining-hall:8080",
}

func SetConfig(c Config) {
	cfg = c
}
