package domain

import (
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type FoodOrder struct {
	CookingDetail
	ItemId  int
	OrderId int
}

type CookDetails struct {
	Rank        int    `json:"rank"`
	Proficiency int64  `json:"proficiency"`
	Name        string `json:"name"`
	CatchPhrase string `json:"catch_phrase"`
}

type Cook struct {
	CookDetails
	Id             int
	Occupation     int64
	SendCookedFood chan<- FoodOrder
	Menu           Menu
}

func NewCook(id int, cookDetails CookDetails, sendChan chan<- FoodOrder, menu Menu) *Cook {
	return &Cook{
		CookDetails:    cookDetails,
		Id:             id,
		Occupation:     0,
		SendCookedFood: sendChan,
		Menu:           menu,
	}
}

func (c *Cook) CanCook(food Food) bool {
	isFree := atomic.LoadInt64(&c.Occupation) < c.Proficiency
	isQualified := food.Complexity <= c.Rank
	return isFree && isQualified
}

func (c *Cook) CookFood(foodOrder FoodOrder) {
	food := c.Menu.Foods[foodOrder.FoodId-1]

	if food.Complexity > c.Rank {
		log.Warn().Int("cook_id", c.Id).Msgf("%s is not qualified to cook %s", c.Name, food.Name)
		atomic.AddInt64(&c.Occupation, -1)
		return
	}

	if atomic.LoadInt64(&c.Occupation) > c.Proficiency {
		log.Warn().Int("cook_id", c.Id).Msgf("%s is busy", c.Name)
		atomic.AddInt64(&c.Occupation, -1)
		return
	}

	preparationTime := time.Duration(food.PreparationTime * cfg.TimeUnit * int(time.Millisecond))
	time.Sleep(preparationTime)

	log.Debug().Int("cook_id", c.Id).Int("food_id", foodOrder.FoodId).Int("order_id", foodOrder.OrderId).Msgf("%s finished cooking %s", c.Name, food.Name)
	c.SendCookedFood <- foodOrder
	atomic.AddInt64(&c.Occupation, -1)
}
