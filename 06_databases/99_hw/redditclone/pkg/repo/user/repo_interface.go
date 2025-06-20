package repo

import (
	"database/sql"

	"redditclone/pkg/session"
)

type UserRepo interface {
	Authorize(login, pass string) (string, error)
	Register(login, pass string) (string, error)
}

type UserMemoryRepository struct {
	db      *sql.DB
	session session.JWTGenerator
	hasher  Hasher
	idgen   IDGenerator
}

type Hasher interface {
	HashPassword(pass string) ([]byte, error)
}

type IDGenerator interface {
	GenerateID() ([]byte, error)
}
