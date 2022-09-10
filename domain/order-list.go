package domain

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	diningHallPath = "http://localhost:8080/distribution"
)

type OrderList struct {
	Distributions     map[int]*Distribution
	ReceiveOrder      <-chan Order
	ReceiveCookedFood chan FoodOrder
	Cooks             []*Cook
	Menu              Menu
}

type CooksDetails struct {
	Cooks []CookDetails
}

func NewOrderList(receiveOrder <-chan Order, menu Menu) *OrderList {
	ol := &OrderList{
		Distributions:     make(map[int]*Distribution),
		ReceiveOrder:      receiveOrder,
		ReceiveCookedFood: make(chan FoodOrder),
		Menu:              menu,
	}

	file, err := os.Open("config/cooks.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening cooks.json")
	}

	byteValue, _ := ioutil.ReadAll(file)
	var cooksDetails CooksDetails
	json.Unmarshal(byteValue, &cooksDetails)

	ol.Cooks = make([]*Cook, len(cooksDetails.Cooks))
	for i, cookDetails := range cooksDetails.Cooks {
		ol.Cooks[i] = NewCook(i, cookDetails, ol.ReceiveCookedFood, ol.Menu)
		log.Info().Int("cook_id", i).Msgf("%s entered the kitchen", cookDetails.Name)
	}

	return ol
}

func (ol *OrderList) Run() {
	go ol.SendFoodOrderToCooks()
	go ol.ReceiveFoodOrderFromCooks()
}

func (ol *OrderList) SendFoodOrderToCooks() {
	for order := range ol.ReceiveOrder {
		ol.Distributions[order.OrderId] = &Distribution{
			Order:          order,
			CookingTime:    time.Now().UnixMilli(),
			CookingDetails: make([]CookingDetail, 0),
			ReceivedItems:  make([]bool, len(order.Items)),
		}

		log.Info().Int("order_id", order.OrderId).Msg("Kitchen received order")

		for i, id := range order.Items {
			food := ol.Menu.Foods[id-1]
			IsFoodOrderSent := false

			for !IsFoodOrderSent {
				for _, cook := range ol.Cooks {
					if cook.CanCook(food) {

						foodOrder := FoodOrder{
							OrderId: order.OrderId,
							ItemId:  i,
							CookingDetail: CookingDetail{
								FoodId: food.Id,
								CookId: cook.Id,
							},
						}

						atomic.AddInt64(&cook.Occupation, 1)
						go cook.CookFood(foodOrder)
						log.Info().Int("order_id", order.OrderId).Int("item_id", i).Int("food_id", food.Id).Int("cook_id", cook.Id).Msgf("%s order assigned to %s", food.Name, cook.Name)

						IsFoodOrderSent = true
						break
					}
				}
			}
		}
	}
}

func (ol *OrderList) ReceiveFoodOrderFromCooks() {
	for foodOrder := range ol.ReceiveCookedFood {

		distribution := ol.Distributions[foodOrder.OrderId]

		if distribution.Order.Items[foodOrder.ItemId] != foodOrder.FoodId {
			log.Warn().Int("order_id", foodOrder.OrderId).Int("item_id", foodOrder.ItemId).Msg("Received wrong food item")
			continue
		}

		if distribution.ReceivedItems[foodOrder.ItemId] {
			log.Warn().Int("order_id", foodOrder.OrderId).Int("item_id", foodOrder.ItemId).Msg("Food item already received")
			continue
		}

		distribution.ReceivedItems[foodOrder.ItemId] = true
		distribution.CookingDetails = append(distribution.CookingDetails, foodOrder.CookingDetail)

		if len(distribution.CookingDetails) == len(distribution.Order.Items) {

			distribution.CookingTime = (time.Now().UnixMilli() - distribution.CookingTime) / int64(timeUnit)

			jsonBody, err := json.Marshal(distribution)
			if err != nil {
				log.Fatal().Err(err).Msg("Error marshalling distribution")
			}
			contentType := "application/json"

			_, err = http.Post(diningHallPath, contentType, bytes.NewReader(jsonBody))
			if err != nil {
				log.Fatal().Err(err).Msg("Error sending distribution to dining hall")
			}

			log.Info().Int("order_id", foodOrder.OrderId).Msg("Distribution sent to dining hall")
			delete(ol.Distributions, foodOrder.OrderId)
		}
	}
}
