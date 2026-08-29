package vehicle

type Vehicle struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	Brand     string `json:"brand"`
	Plate     string `json:"plate"`
	Color     string `json:"color"`
	Year      int    `json:"year"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
