package domain

import "time"

const (
	timeUnit = 250
)

type Order struct {
	OrderId    int       `json:"order_id"`
	TableId    int       `json:"table_id"`
	WaiterId   int       `json:"waiter_id"`
	Items      []int     `json:"items"`
	Priority   int       `json:"priority"`
	MaxWait    float64   `json:"max_wait"`
	PickUpTime time.Time `json:"pick_up_time"`
}
