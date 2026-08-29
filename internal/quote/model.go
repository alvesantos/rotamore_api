package quote

type Quote struct {
	ID          string   `json:"id"`
	UserID      string   `json:"user_id"`
	Category    string   `json:"category"` // 'transfer' | 'passeio'
	Pickup      string   `json:"pickup"`
	Destination string   `json:"destination"`
	Stops       []string `json:"stops"`
	Price       float64  `json:"price"`
	CreatedAt   string   `json:"created_at"`
}
