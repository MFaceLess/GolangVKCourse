package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	customError "redditclone/pkg/errors"
	repo "redditclone/pkg/repo/user"
)

type brokenWrite struct{}

func (bw *brokenWrite) Header() http.Header {
	return http.Header{}
}

func (bw *brokenWrite) Write([]byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

func (bw *brokenWrite) WriteHeader(statusCode int) {}

func TestUserHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := repo.NewMockUserRepo(ctrl)
	handler := &UserHandler{
		UserRepo: userRepo,
	}

	userCredentials := `{"username": "testuser", "password": "password123"}`

	t.Run("successful login", func(t *testing.T) {
		userRepo.EXPECT().Authorize("testuser", "password123").Return("valid_token", nil)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(userCredentials)))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		var response SignInResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)
		assert.Equal(t, "valid_token", response.Token)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		userRepo.EXPECT().Authorize("testuser", "wrongpassword").Return("", repo.ErrNotFoundUser)

		body := `{"username":"testuser", "password":"wrongpassword"}`
		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(body)))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected status code 401, got %d", resp.StatusCode)
		}

		var response struct {
			Message string `json:"message"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)

		assert.Nil(t, err)
		assert.Equal(t, response.Message, "invalid password")
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidCredentials := `{"username": "testuser", "password":}`

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(invalidCredentials)))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("authorization error", func(t *testing.T) {
		userRepo.EXPECT().Authorize("testuser", "password123").Return("", errors.New("authorization failed"))

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(userCredentials)))
		w := httptest.NewRecorder()

		handler.Login(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("json encode error after unauthorized", func(t *testing.T) {
		userRepo.EXPECT().Authorize("testuser", "wrongpassword").Return("", repo.ErrNotFoundUser)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(`{"username":"testuser","password":"wrongpassword"}`)))
		w := &brokenWrite{}

		handler.Login(w, req)
	})

	t.Run("json encode error after success", func(t *testing.T) {
		userRepo.EXPECT().Authorize("testuser", "password123").Return("sometoken", nil)

		req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte(`{"username":"testuser","password":"password123"}`)))
		w := &brokenWrite{}

		handler.Login(w, req)
	})
}

func TestUserHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := repo.NewMockUserRepo(ctrl)
	handler := &UserHandler{
		UserRepo: userRepo,
	}

	userCredentials := `{"username": "testuser", "password": "password123"}`

	t.Run("invalid json", func(t *testing.T) {
		invalidCredentials := `{"username": "testuser", "password":}`

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(invalidCredentials)))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusInternalServerError)
	})

	t.Run("successful login", func(t *testing.T) {
		userRepo.EXPECT().Register("testuser", "password123").Return("", repo.ErrUserAlreadyExists)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(userCredentials)))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity)

		var response struct {
			Errors []customError.ErrorResponse `json:"errors"`
		}

		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Nil(t, err)

		assert.Equal(t, "body", response.Errors[0].Location)
		assert.Equal(t, "already exists", response.Errors[0].Msg)
		assert.Equal(t, "username", response.Errors[0].Param)
		assert.Equal(t, "testuser", response.Errors[0].Value)
	})

	t.Run("register error", func(t *testing.T) {
		userRepo.EXPECT().Register("testuser", "password123").Return("", fmt.Errorf("unknown error"))

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(userCredentials)))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusBadRequest)
	})

	t.Run("json error after register", func(t *testing.T) {
		userRepo.EXPECT().Register("testuser", "password123").Return("token", nil)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(userCredentials)))

		handler.Register(&brokenWrite{}, req)
	})

	t.Run("success", func(t *testing.T) {
		userRepo.EXPECT().Register("testuser", "password123").Return("token", nil)

		req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte(userCredentials)))
		w := httptest.NewRecorder()

		handler.Register(w, req)

		resp := w.Result()
		assert.Equal(t, resp.StatusCode, http.StatusCreated)

		var response SignInResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		assert.Nil(t, err)
		assert.Equal(t, "token", response.Token)
	})
}
