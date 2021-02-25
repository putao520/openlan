package schema

type User struct {
	Alias    string `json:"alias,omitempty"`
	Role     string `json:"role,omitempty"` // admin, guest or other
	Name     string `json:"name"`
	Password string `json:"password"`
	Token    string `json:"token,omitempty"`
	Network  string `json:"network"`
}

