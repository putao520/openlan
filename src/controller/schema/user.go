package schema

type User struct {
	Role     string `json:"role"` // admin, guest or other
	Name     string `json:"name"`
	Password string `json:"password"`
}
