package user

type User struct {
	Login    string `json:"username"`
	ID       string `json:"id"`
	Password []byte
}
