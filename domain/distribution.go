package domain

type Distribution struct {
	Order

	CookingTime    int             `json:"cooking_time"`
	CookingDetails []CookingDetail `json:"cooking_details"`
}

type CookingDetail struct {
	FoodId int `json:"food_id"`
	CookId int `json:"cook_id"`
}
