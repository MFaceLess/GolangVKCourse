package handlers

import (
	"encoding/json"
	"net/http"

	"redditclone/pkg/errors"
	repo "redditclone/pkg/repo/user"
)

type UserHandler struct {
	UserRepo repo.UserRepo
}

type SignInResponse struct {
	Token string `json:"token"`
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var userCredentials struct {
		Login    string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&userCredentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := h.UserRepo.Authorize(userCredentials.Login, userCredentials.Password)
	if err != nil {
		if err == repo.ErrNotFoundUser {
			w.WriteHeader(http.StatusUnauthorized)

			var response struct {
				Message string `json:"message"`
			}

			response.Message = "invalid password"

			if err = json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	response := SignInResponse{
		Token: tokenString,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var userCredentials struct {
		Login    string `json:"username"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&userCredentials)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tokenString, err := h.UserRepo.Register(userCredentials.Login, userCredentials.Password)
	if err != nil {
		if err == repo.ErrUserAlreadyExists {
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors.ErrorJSON(w, "body", "username", userCredentials.Login, "already exists")
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	response := SignInResponse{
		Token: tokenString,
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
