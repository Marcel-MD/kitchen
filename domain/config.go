package domain

type Config struct {
	TimeUnit             int `json:"time_unit"`
	NrOfTables           int `json:"nr_of_tables"`
	MaxOrderItemsCount   int `json:"max_order_items_count"`
	CtxSwitchFactor      int `json:"ctx_switch_factor"`
	NrOfConcurrentOrders int `json:"nr_of_concurrent_orders"`

	DiningHallUrl string `json:"dining_hall_url"`
	KitchenPort   string `json:"kitchen_port"`
}

var cfg Config = Config{
	TimeUnit:             250,
	NrOfTables:           10,
	MaxOrderItemsCount:   10,
	CtxSwitchFactor:      3,
	NrOfConcurrentOrders: 2,

	DiningHallUrl: "http://dining-hall:8080",
	KitchenPort:   "8081",
}

func SetConfig(c Config) {
	cfg = c
}
