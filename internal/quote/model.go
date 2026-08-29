package quote

type Quote struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Pickup      string  `json:"pickup"`
	Destination string  `json:"destination"`
	Price       float64 `json:"price"`
	CreatedAt   string  `json:"created_at"`
}
