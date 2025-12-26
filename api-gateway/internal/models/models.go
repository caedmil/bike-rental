package models

type RentResponse struct {
	RentID    string `json:"rent_id"`
	UserID    string `json:"user_id"`
	BikeID    string `json:"bike_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
}

type BikesList struct {
	Bikes []Bike `json:"bikes"`
}

type Bike struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Location string `json:"location"`
}

