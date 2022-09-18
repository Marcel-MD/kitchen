package domain

import (
	"bytes"
	"container/heap"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
)

type OrderList struct {
	Distributions      map[int]*Distribution
	ReceiveOrder       <-chan Order
	OrderQueueChan     chan Order
	ReceiveCookedFood  chan FoodOrder
	Cooks              []*Cook
	Menu               Menu
	Apparatuses        map[string]*Apparatus
	Queue              PriorityQueue
	NrProcessingOrders int64
}

type cooksDetails struct {
	Cooks []CookDetails
}

func NewOrderList(receiveOrder <-chan Order) *OrderList {
	ol := &OrderList{
		Distributions:      make(map[int]*Distribution),
		ReceiveOrder:       receiveOrder,
		OrderQueueChan:     make(chan Order),
		ReceiveCookedFood:  make(chan FoodOrder),
		Menu:               GetMenu(),
		Apparatuses:        GetApparatusesMap(),
		Queue:              make(PriorityQueue, 0, 10),
		NrProcessingOrders: 0,
	}

	file, err := os.Open("config/cooks.json")
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening cooks.json")
	}

	byteValue, _ := ioutil.ReadAll(file)
	var cds cooksDetails
	json.Unmarshal(byteValue, &cds)

	ol.Cooks = make([]*Cook, len(cds.Cooks))
	for i, cookDetails := range cds.Cooks {
		ol.Cooks[i] = NewCook(i, cookDetails, ol.ReceiveCookedFood, ol.Menu, ol.Apparatuses)
		log.Debug().Int("cook_id", i).Msgf("%s entered the kitchen", cookDetails.Name)
	}

	sort.Slice(ol.Cooks, func(i, j int) bool {
		return ol.Cooks[i].Proficiency < ol.Cooks[j].Proficiency
	})

	return ol
}

func (ol *OrderList) Run() {
	go ol.manageQueue()
	go ol.sendFoodOrderToCooks()
	go ol.receiveFoodOrderFromCooks()
}

func (ol *OrderList) manageQueue() {
	for {
		select {
		case order := <-ol.ReceiveOrder:
			log.Info().Int("order_id", order.OrderId).Msg("Kitchen received order")
			heap.Push(&ol.Queue, &Item{Order: order})

		default:
			if atomic.LoadInt64(&ol.NrProcessingOrders) == 0 && len(ol.Queue) > 0 {
				item := heap.Pop(&ol.Queue).(*Item)
				atomic.AddInt64(&ol.NrProcessingOrders, 1)
				ol.OrderQueueChan <- item.Order
			}
		}
	}
}

func (ol *OrderList) sendFoodOrderToCooks() {
	for order := range ol.OrderQueueChan {
		ol.Distributions[order.OrderId] = &Distribution{
			Order:          order,
			CookingTime:    time.Now().UnixMilli(),
			CookingDetails: make([]CookingDetail, 0),
			ReceivedItems:  make([]bool, len(order.Items)),
		}

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
						log.Debug().Int("order_id", order.OrderId).Int("item_id", i).Int("food_id", food.Id).Int("cook_id", cook.Id).Msgf("%s order assigned to %s", food.Name, cook.Name)

						IsFoodOrderSent = true
						break
					}
				}
			}
		}

		atomic.AddInt64(&ol.NrProcessingOrders, -1)
	}
}

func (ol *OrderList) receiveFoodOrderFromCooks() {
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
			ol.sendDistributionToDiningHall(*distribution)
		}
	}
}

func (ol *OrderList) sendDistributionToDiningHall(distribution Distribution) {
	distribution.CookingTime = (time.Now().UnixMilli() - distribution.CookingTime) / int64(cfg.TimeUnit)

	jsonBody, err := json.Marshal(distribution)
	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling distribution")
	}
	contentType := "application/json"

	_, err = http.Post(cfg.DiningHallUrl+"/distribution", contentType, bytes.NewReader(jsonBody))
	if err != nil {
		log.Fatal().Err(err).Msg("Error sending distribution to dining hall")
	}

	log.Info().Int("order_id", distribution.OrderId).Int64("cooking_time", distribution.CookingTime).Msg("Distribution sent to dining hall")
	delete(ol.Distributions, distribution.OrderId)
}
