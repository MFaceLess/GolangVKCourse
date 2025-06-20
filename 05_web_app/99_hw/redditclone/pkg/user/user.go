package user

type User struct {
	Login    string `json:"username"`
	ID       string `json:"id"`
	password []byte
}

type UserRepo interface {
	Authorize(login, pass string) (string, error)
	Register(login, pass string) (string, error)
}
