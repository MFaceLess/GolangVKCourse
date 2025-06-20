package user

import (
	"crypto/rand"
	"errors"
	"fmt"

	"redditclone/pkg/session"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFoundUser             = errors.New("User not found")
	ErrUserAlreadyExists        = errors.New("User already exists")
	ErrCantGenerateHashPassword = errors.New("error hash password")
)

type UserMemoryRepository struct {
	data map[string]*User
	*sync.RWMutex
}

func NewMemoryRepo() *UserMemoryRepository {
	return &UserMemoryRepository{
		data:    map[string]*User{},
		RWMutex: &sync.RWMutex{},
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (string, error) {
	repo.RLock()
	u, ok := repo.data[login]
	if !ok {
		repo.RUnlock()
		return "", ErrNotFoundUser
	}
	repo.RUnlock()

	if err := bcrypt.CompareHashAndPassword(u.password, []byte(pass)); err != nil {
		return "", ErrNotFoundUser
	}

	tokenString, err := session.GenerateJWT(login, u.ID)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (repo *UserMemoryRepository) Register(login, pass string) (string, error) {
	repo.RLock()
	if _, ok := repo.data[login]; ok {
		repo.RUnlock()
		return "", ErrUserAlreadyExists
	}
	repo.RUnlock()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrCantGenerateHashPassword
	}

	randID := make([]byte, 16)
	if _, err = rand.Read(randID); err != nil {
		return "", err
	}

	newUser := &User{password: hashedPassword, Login: login, ID: fmt.Sprintf("%x", randID)}

	repo.Lock()
	repo.data[login] = newUser
	repo.Unlock()

	randID = make([]byte, 16)
	if _, err = rand.Read(randID); err != nil {
		return "", err
	}

	tokenString, err := session.GenerateJWT(login, fmt.Sprintf("%x", randID))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
