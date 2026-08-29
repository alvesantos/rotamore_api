package ride

type Ride struct {
	ID              string   `json:"id"`
	UserID          string   `json:"user_id"`
	VehicleID       *string  `json:"vehicle_id,omitempty"`
	VehicleName     string   `json:"vehicle_name,omitempty"`
	VehicleBrand    string   `json:"vehicle_brand,omitempty"`
	VehiclePlate    string   `json:"vehicle_plate,omitempty"`
	VehicleColor    string   `json:"vehicle_color,omitempty"`
	Category        string   `json:"category"` // 'transfer' | 'passeio'
	CustomerName    string   `json:"customer_name"`
	CustomerPhone   string   `json:"customer_phone"`
	PassengersCount int      `json:"passengers_count"`
	Pickup          string   `json:"pickup"`
	Destination     string   `json:"destination"`
	Stops           []string `json:"stops"` // List of stops/destinations for Passeio
	Notes           string   `json:"notes"`
	RideDate        string   `json:"ride_date"`
	RideTime        string   `json:"ride_time"`
	Price           float64  `json:"price"`
	Status          string   `json:"status"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}
