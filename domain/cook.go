package domain

import (
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type FoodOrder struct {
	CookingDetail
	ItemId                   int
	OrderId                  int
	Food                     Food
	RemainingPreparationTime int
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
	FoodOrderChan  chan<- FoodOrder
	Menu           Menu
	Apparatuses    map[string]*Apparatus
}

func NewCook(id int, cookDetails CookDetails, sendChan chan<- FoodOrder, foodOrderChan chan<- FoodOrder, menu Menu, apparatuses map[string]*Apparatus) *Cook {
	return &Cook{
		CookDetails:    cookDetails,
		Id:             id,
		Occupation:     0,
		SendCookedFood: sendChan,
		FoodOrderChan:  foodOrderChan,
		Menu:           menu,
		Apparatuses:    apparatuses,
	}
}

func (c *Cook) CanCook(food Food) bool {
	isFree := atomic.LoadInt64(&c.Occupation) < c.Proficiency
	isQualified := food.Complexity <= c.Rank
	return isFree && isQualified
}

func (c *Cook) CookFood(fo FoodOrder) {
	food := fo.Food

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

	if food.CookingApparatus != "" {
		apparatus := c.Apparatuses[food.CookingApparatus]
		if apparatus != nil {
			for {
				if apparatus.Use() {
					go c.cookFoodWithApparatus(fo, apparatus)
					atomic.AddInt64(&c.Occupation, -1)
					return
				}
				time.Sleep(time.Duration(cfg.TimeUnit) * time.Millisecond)
			}
		}
	}

	ctxSwitchTime := food.PreparationTime / cfg.CtxSwitchFactor
	if fo.RemainingPreparationTime-ctxSwitchTime > 1 {
		preparationTime := time.Duration(ctxSwitchTime * cfg.TimeUnit * int(time.Millisecond))
		time.Sleep(preparationTime)
		fo.RemainingPreparationTime -= ctxSwitchTime
		c.FoodOrderChan <- fo
		atomic.AddInt64(&c.Occupation, -1)
		return
	}

	preparationTime := time.Duration(fo.RemainingPreparationTime * cfg.TimeUnit * int(time.Millisecond))
	time.Sleep(preparationTime)
	fo.RemainingPreparationTime = 0

	log.Debug().Int("cook_id", c.Id).Int("food_id", fo.FoodId).Int("order_id", fo.OrderId).Msgf("%s finished cooking %s", c.Name, food.Name)
	c.SendCookedFood <- fo
	atomic.AddInt64(&c.Occupation, -1)
}

func (c *Cook) cookFoodWithApparatus(fo FoodOrder, apparatus *Apparatus) {
	food := fo.Food

	ctxSwitchTime := food.PreparationTime / cfg.CtxSwitchFactor
	if fo.RemainingPreparationTime-ctxSwitchTime > 0 {
		preparationTime := time.Duration(ctxSwitchTime * cfg.TimeUnit * int(time.Millisecond))
		time.Sleep(preparationTime)
		fo.RemainingPreparationTime -= ctxSwitchTime
		c.FoodOrderChan <- fo
		apparatus.Release()
		return
	}

	preparationTime := time.Duration(fo.RemainingPreparationTime * cfg.TimeUnit * int(time.Millisecond))
	time.Sleep(preparationTime)
	fo.RemainingPreparationTime = 0

	log.Debug().Int("cook_id", c.Id).Int("food_id", fo.FoodId).Int("order_id", fo.OrderId).Msgf("%s finished cooking %s with %s", c.Name, food.Name, apparatus.Name)
	c.SendCookedFood <- fo
	apparatus.Release()
}
