package user

type UserType string

const (
	TypeDriver   UserType = "driver"
	TypeCustomer UserType = "customer"
	TypeAdmin    UserType = "admin"
)

type User struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	LastName     string   `json:"lastname"`
	Phone        string   `json:"phone"`
	Type         UserType `json:"type"`
	Email        string   `json:"email"`
	Document     string   `json:"document"`
	PasswordHash string   `json:"-"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}
