package handlers

import (
	"encoding/json"
	"net/http"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/session"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type PostHandler struct {
	Logger   *zap.SugaredLogger
	PostRepo post.PostRepo
}

func (h *PostHandler) GetAllPosts(w http.ResponseWriter, r *http.Request) {
	allPosts := h.PostRepo.GetAll()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(allPosts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	const claimsCtxKey middleware.ContextKey = "claims"
	claims, ok := r.Context().Value(claimsCtxKey).(*session.Claims)
	if !ok {
		http.Error(w, "Невалидный токен", http.StatusBadRequest)
		return
	}

	var postData post.DataPost
	err := json.NewDecoder(r.Body).Decode(&postData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post := h.PostRepo.Add(postData, claims.User.Username, claims.User.UserID)
	if post == nil {
		http.Error(w, "Ошибка при добавлении поста", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *PostHandler) GetPostByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	categoryName := vars["CATEGORY_NAME"]

	posts := h.PostRepo.GetByCategory(categoryName)
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["POST_ID"]

	post := h.PostRepo.GetByID(id)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) GetPostsByUsername(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["USER_LOGIN"]

	posts := h.PostRepo.GetPostsByUsername(username)
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]

	const claimsCtxKey middleware.ContextKey = "claims"
	claims, ok := r.Context().Value(claimsCtxKey).(*session.Claims)
	if !ok {
		http.Error(w, "Невалидный токен", http.StatusBadRequest)
		return
	}

	var response struct {
		Message string `json:"message"`
	}

	err := h.PostRepo.DeletePostByID(postID, claims.User.Username)
	if err != nil {
		if err == post.ErrAccessDenied {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		response.Message = err.Error()
	}

	if response.Message == "" {
		w.WriteHeader(http.StatusOK)
		response.Message = "success"
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PostHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]

	const claimsCtxKey middleware.ContextKey = "claims"
	claims, ok := r.Context().Value(claimsCtxKey).(*session.Claims)
	if !ok {
		http.Error(w, "Невалидный токен", http.StatusBadRequest)
		return
	}

	var postData post.DataComment
	err := json.NewDecoder(r.Body).Decode(&postData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := h.PostRepo.AddComment(postID, postData.Comment, claims.User.Username, claims.User.UserID)

	if err != nil {
		var response struct {
			Message string `json:"message"`
		}

		response.Message = err.Error()
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	commentID := vars["COMMENT_ID"]

	const claimsCtxKey middleware.ContextKey = "claims"
	claims, ok := r.Context().Value(claimsCtxKey).(*session.Claims)
	if !ok {
		http.Error(w, "Невалидный токен", http.StatusBadRequest)
		return
	}

	var response struct {
		Message string `json:"message"`
	}

	post, err := h.PostRepo.DeleteComment(postID, commentID, claims.User.Username)
	if err != nil {
		response.Message = err.Error()
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *PostHandler) VotePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	voteType := vars["VOTE_TYPE"]

	const claimsCtxKey middleware.ContextKey = "claims"
	claims, ok := r.Context().Value(claimsCtxKey).(*session.Claims)
	if !ok {
		http.Error(w, "Невалидный токен", http.StatusBadRequest)
		return
	}

	var response struct {
		Message string `json:"message"`
	}

	var post *post.Post
	var err error

	switch voteType {
	case "upvote":
		post, err = h.PostRepo.VotePost(postID, claims.User.UserID, 1)
	case "downvote":
		post, err = h.PostRepo.VotePost(postID, claims.User.UserID, -1)
	case "unvote":
		post, err = h.PostRepo.VotePost(postID, claims.User.UserID, 0)

	}

	if err != nil {
		response.Message = err.Error()
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
