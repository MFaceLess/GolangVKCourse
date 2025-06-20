package repo

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"redditclone/pkg/session"
	"redditclone/pkg/user"
)

var (
	ErrNotFoundUser             = errors.New("user not found")
	ErrUserAlreadyExists        = errors.New("user already exists")
	ErrCantGenerateHashPassword = errors.New("error hash password")
)

type BcryptHasher struct{}

func (BcryptHasher) HashPassword(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
}

type RandomIDGenerator struct{}

func (RandomIDGenerator) GenerateID() ([]byte, error) {
	id := make([]byte, 16)
	_, err := rand.Read(id)
	return id, err
}

func NewMemoryRepo(db *sql.DB, jwtGen session.JWTGenerator) *UserMemoryRepository {
	return &UserMemoryRepository{
		db:      db,
		session: jwtGen,
		hasher:  BcryptHasher{},
		idgen:   RandomIDGenerator{},
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (string, error) {
	var (
		userID         string
		hashedPassword []byte
	)

	err := repo.db.QueryRow(
		"SELECT id, password FROM users WHERE login = ?",
		login,
	).Scan(&userID, &hashedPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFoundUser
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(pass)); err != nil {
		return "", ErrNotFoundUser
	}

	return repo.session.GenerateJWT(login, userID)
}

func (repo *UserMemoryRepository) Register(login, pass string) (string, error) {
	var exists bool
	err := repo.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE login = ?)",
		login,
	).Scan(&exists)

	if err != nil {
		return "", err
	}

	if exists {
		return "", ErrUserAlreadyExists
	}

	hashedPassword, err := repo.hasher.HashPassword(pass)
	if err != nil {
		return "", ErrCantGenerateHashPassword
	}

	randID, err := repo.idgen.GenerateID()
	if err != nil {
		return "", err
	}

	newUser := &user.User{Password: hashedPassword, Login: login, ID: fmt.Sprintf("%x", randID)}

	_, err = repo.db.Exec(
		"INSERT INTO users (id, login, password) VALUES (?, ?, ?)",
		newUser.ID,
		newUser.Login,
		newUser.Password,
	)

	if err != nil {
		return "", err
	}

	tokenString, err := repo.session.GenerateJWT(login, newUser.ID)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
