package domain

type Distribution struct {
	Order

	CookingTime    int64           `json:"cooking_time"`
	CookingDetails []CookingDetail `json:"cooking_details"`
	ReceivedItems  []bool          `json:"-"`
}

type CookingDetail struct {
	FoodId int `json:"food_id"`
	CookId int `json:"cook_id"`
}
