package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	repo "redditclone/pkg/repo/post"
	"redditclone/pkg/session"
)

func TestGetAllPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{
		PostRepo: postRepo,
	}

	t.Run("success", func(t *testing.T) {
		expectedPosts := []post.Post{
			{ID: primitive.NewObjectID(), Title: "Post 1"},
			{ID: primitive.NewObjectID(), Title: "Post 2"},
		}

		postRepo.EXPECT().GetAll().Return(expectedPosts, nil)

		req := httptest.NewRequest("GET", "/posts", nil)
		w := httptest.NewRecorder()

		handler.GetAllPosts(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var got []post.Post
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expectedPosts, got)
	})

	t.Run("repository error", func(t *testing.T) {
		postRepo.EXPECT().GetAll().Return(nil, errors.New("db error"))

		req := httptest.NewRequest("GET", "/posts", nil)
		w := httptest.NewRecorder()

		handler.GetAllPosts(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("json error", func(t *testing.T) {
		expectedPosts := []post.Post{
			{ID: primitive.NewObjectID(), Title: "Post 1"},
			{ID: primitive.NewObjectID(), Title: "Post 2"},
		}

		postRepo.EXPECT().GetAll().Return(expectedPosts, nil)

		req := httptest.NewRequest("GET", "/posts", nil)
		w := &brokenWrite{}

		handler.GetAllPosts(w, req)
	})
}

func TestAddPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: postRepo}

	claims := &session.Claims{
		User: session.User{
			Username: "testuser",
			UserID:   "123",
		},
	}

	ctxKey := middleware.ContextKey("claims")

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts", nil)
		w := httptest.NewRecorder()

		handler.AddPost(w, req.WithContext(context.Background()))

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("bad json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts", strings.NewReader("{bad json"))
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		handler.AddPost(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("repo error", func(t *testing.T) {
		data := post.DataPost{Title: "Test"}
		body, err := json.Marshal(data)
		if err != nil {
			return
		}

		req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		postRepo.EXPECT().Add(data, "testuser", "123").Return(post.Post{}, errors.New("failed"))

		handler.AddPost(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("encode error", func(t *testing.T) {
		data := post.DataPost{Title: "Test"}
		body, err := json.Marshal(data)
		if err != nil {
			return
		}
		req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := &brokenWrite{}

		postRepo.EXPECT().Add(data, "testuser", "123").Return(post.Post{Title: "OK"}, nil)

		handler.AddPost(w, req)
	})

	t.Run("success", func(t *testing.T) {
		data := post.DataPost{Title: "Test"}
		body, err := json.Marshal(data)
		if err != nil {
			return
		}
		req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		expected := post.Post{Title: "Test"}

		postRepo.EXPECT().Add(data, "testuser", "123").Return(expected, nil)

		handler.AddPost(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var got post.Post
		err = json.NewDecoder(resp.Body).Decode(&got)
		if err != nil {
			return
		}
		assert.Equal(t, expected.Title, got.Title)
	})
}

func TestGetPostByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: postRepo}

	t.Run("internal error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/category/programming", nil)
		req = mux.SetURLVars(req, map[string]string{"CATEGORY_NAME": "programming"})
		w := httptest.NewRecorder()

		postRepo.EXPECT().GetByCategory("programming").Return(nil, errors.New("db error"))

		handler.GetPostByCategory(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("encode error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/category/bugs", nil)
		req = mux.SetURLVars(req, map[string]string{"CATEGORY_NAME": "bugs"})
		w := &brokenWrite{}

		postRepo.EXPECT().GetByCategory("bugs").Return([]post.Post{{Title: "bug"}}, nil)

		handler.GetPostByCategory(w, req)
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/category/news", nil)
		req = mux.SetURLVars(req, map[string]string{"CATEGORY_NAME": "news"})
		w := httptest.NewRecorder()

		expected := []post.Post{{Title: "latest"}}

		postRepo.EXPECT().GetByCategory("news").Return(expected, nil)

		handler.GetPostByCategory(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got []*post.Post
		err := json.NewDecoder(resp.Body).Decode(&got)
		if err != nil {
			return
		}
		assert.Equal(t, expected[0].Title, got[0].Title)
	})
}

func TestGetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: postRepo}

	t.Run("internal error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/post/123", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "123"})
		w := httptest.NewRecorder()

		postRepo.EXPECT().GetByID("123").Return(post.Post{}, errors.New("db error"))

		handler.GetPostByID(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("encode error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/post/456", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "456"})
		w := &brokenWrite{}

		postRepo.EXPECT().GetByID("456").Return(post.Post{Title: "Test"}, nil)

		handler.GetPostByID(w, req)
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/post/789", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "789"})
		w := httptest.NewRecorder()

		expected := post.Post{Title: "Hello"}

		postRepo.EXPECT().GetByID("789").Return(expected, nil)

		handler.GetPostByID(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got post.Post
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expected.Title, got.Title)
	})
}

func TestGetPostsByUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: mockRepo}

	t.Run("internal error from repo", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/testuser/posts", nil)
		req = mux.SetURLVars(req, map[string]string{"USER_LOGIN": "testuser"})
		w := httptest.NewRecorder()

		mockRepo.EXPECT().GetPostsByUsername("testuser").Return([]post.Post{}, errors.New("db error"))

		handler.GetPostsByUsername(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("encoding error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/john/posts", nil)
		req = mux.SetURLVars(req, map[string]string{"USER_LOGIN": "john"})
		w := &brokenWrite{}

		mockRepo.EXPECT().GetPostsByUsername("john").Return([]post.Post{{Title: "Title"}}, nil)

		handler.GetPostsByUsername(w, req)
	})

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/user/alice/posts", nil)
		req = mux.SetURLVars(req, map[string]string{"USER_LOGIN": "alice"})
		w := httptest.NewRecorder()

		expected := []post.Post{{Title: "First Post"}, {Title: "Second Post"}}
		mockRepo.EXPECT().GetPostsByUsername("alice").Return(expected, nil)

		handler.GetPostsByUsername(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got []post.Post
		err := json.NewDecoder(resp.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expected, got)
	})
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: mockRepo}

	user := &session.User{Username: "alice", UserID: "123"}
	claims := &session.Claims{User: *user}
	ctxKey := middleware.ContextKey("claims")

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		mockRepo.EXPECT().DeletePostByID("abc", "alice").Return(nil)

		handler.DeletePost(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]string
		err := json.NewDecoder(resp.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "success", body["message"])
	})

	t.Run("access denied", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		mockRepo.EXPECT().DeletePostByID("abc", "alice").Return(post.ErrAccessDenied)

		handler.DeletePost(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var body map[string]string
		err := json.NewDecoder(w.Body).Decode(&body)
		if err != nil {
			return
		}
		assert.Contains(t, body["message"], "у вас нет прав на данное действие")
	})

	t.Run("not found", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		mockRepo.EXPECT().DeletePostByID("abc", "alice").Return(errors.New("not found"))

		handler.DeletePost(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var body map[string]string
		err := json.NewDecoder(w.Body).Decode(&body)
		if err != nil {
			return
		}
		assert.Equal(t, "not found", body["message"])
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		w := httptest.NewRecorder()

		handler.DeletePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Невалидный токен")
	})

	t.Run("encode error", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		badWriter := &brokenWrite{}

		mockRepo.EXPECT().DeletePostByID("abc", "alice").Return(nil)

		handler.DeletePost(badWriter, req)
	})
}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: mockRepo}

	ctxKey := middleware.ContextKey("claims")
	claims := &session.Claims{
		User: session.User{Username: "alice", UserID: "123"},
	}

	t.Run("success", func(t *testing.T) {
		comment := post.DataComment{Comment: "Nice post"}
		body, err := json.Marshal(comment)
		if err != nil {
			return
		}

		req := httptest.NewRequest("POST", "/posts/abc/comments", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		expectedPost := post.Post{Title: "test"}
		mockRepo.EXPECT().AddComment("abc", "Nice post", "alice", "123").Return(expectedPost, nil)

		handler.AddComment(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got post.Post
		err = json.NewDecoder(w.Body).Decode(&got)
		if err != nil {
			return
		}
		assert.Equal(t, expectedPost.Title, got.Title)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/comments", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		w := httptest.NewRecorder()

		handler.AddComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Невалидный токен")
	})

	t.Run("bad json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/comments", bytes.NewBufferString("{invalid json"))
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		handler.AddComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("repo error", func(t *testing.T) {
		comment := post.DataComment{Comment: "Fails"}
		body, err := json.Marshal(comment)
		if err != nil {
			return
		}

		req := httptest.NewRequest("POST", "/posts/abc/comments", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		mockRepo.EXPECT().
			AddComment("abc", "Fails", "alice", "123").
			Return(post.Post{}, errors.New("something went wrong"))

		handler.AddComment(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err = json.NewDecoder(w.Body).Decode(&resp)
		if err != nil {
			return
		}
		assert.Equal(t, "something went wrong", resp["message"])
	})

	t.Run("json encode error on failure", func(t *testing.T) {
		comment := post.DataComment{Comment: "Fails"}
		body, err := json.Marshal(comment)
		if err != nil {
			return
		}

		req := httptest.NewRequest("POST", "/posts/abc/comments", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		mockRepo.EXPECT().
			AddComment("abc", "Fails", "alice", "123").
			Return(post.Post{}, errors.New("fail"))

		errWriter := &brokenWrite{}
		handler.AddComment(errWriter, req)
	})

	t.Run("json encode error on success", func(t *testing.T) {
		comment := post.DataComment{Comment: "All good"}
		body, err := json.Marshal(comment)
		if err != nil {
			return
		}

		req := httptest.NewRequest("POST", "/posts/abc/comments", bytes.NewReader(body))
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		mockRepo.EXPECT().
			AddComment("abc", "All good", "alice", "123").
			Return(post.Post{Title: "yay"}, nil)

		handler.AddComment(&brokenWrite{}, req)
	})
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: mockRepo}

	ctxKey := middleware.ContextKey("claims")
	claims := &session.Claims{
		User: session.User{Username: "alice", UserID: "123"},
	}

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc/comments/xyz", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "COMMENT_ID": "xyz"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		expectedPost := post.Post{Title: "updated"}
		mockRepo.EXPECT().DeleteComment("abc", "xyz", "alice").Return(expectedPost, nil)

		handler.DeleteComment(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got post.Post
		err := json.NewDecoder(w.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expectedPost.Title, got.Title)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc/comments/xyz", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "COMMENT_ID": "xyz"})
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Невалидный токен")
	})

	t.Run("repo error", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc/comments/xyz", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "COMMENT_ID": "xyz"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		mockRepo.EXPECT().
			DeleteComment("abc", "xyz", "alice").
			Return(post.Post{}, errors.New("fail to delete"))

		handler.DeleteComment(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var res map[string]string
		err := json.NewDecoder(w.Body).Decode(&res)
		if err != nil {
			return
		}
		assert.Equal(t, "fail to delete", res["message"])
	})

	t.Run("json encode error on failure", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc/comments/xyz", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "COMMENT_ID": "xyz"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		mockRepo.EXPECT().
			DeleteComment("abc", "xyz", "alice").
			Return(post.Post{}, errors.New("something failed"))

		handler.DeleteComment(&brokenWrite{}, req)
	})

	t.Run("json encode error on success", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/posts/abc/comments/xyz", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "COMMENT_ID": "xyz"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		mockRepo.EXPECT().
			DeleteComment("abc", "xyz", "alice").
			Return(post.Post{Title: "ok"}, nil)

		handler.DeleteComment(&brokenWrite{}, req)
	})
}

func TestVotePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := repo.NewMockPostRepo(ctrl)
	handler := &PostHandler{PostRepo: mockRepo}

	ctxKey := middleware.ContextKey("claims")
	claims := &session.Claims{
		User: session.User{Username: "alice", UserID: "123"},
	}

	t.Run("upvote success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/upvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "upvote"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		expectedPost := post.Post{Title: "upvoted"}
		mockRepo.EXPECT().VotePost("abc", "123", 1).Return(expectedPost, nil)

		handler.VotePost(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got post.Post
		err := json.NewDecoder(w.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expectedPost.Title, got.Title)
	})

	t.Run("downvote success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/downvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "downvote"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		expectedPost := post.Post{Title: "downvoted"}
		mockRepo.EXPECT().VotePost("abc", "123", -1).Return(expectedPost, nil)

		handler.VotePost(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got post.Post
		err := json.NewDecoder(w.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expectedPost.Title, got.Title)
	})

	t.Run("unvote success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/unvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "unvote"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		expectedPost := post.Post{Title: "unvoted"}
		mockRepo.EXPECT().VotePost("abc", "123", 0).Return(expectedPost, nil)

		handler.VotePost(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var got post.Post
		err := json.NewDecoder(w.Body).Decode(&got)
		assert.NoError(t, err)
		assert.Equal(t, expectedPost.Title, got.Title)
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/upvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "upvote"})
		w := httptest.NewRecorder()

		handler.VotePost(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Невалидный токен")
	})

	t.Run("repo error", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/upvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "upvote"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))
		w := httptest.NewRecorder()

		mockRepo.EXPECT().VotePost("abc", "123", 1).Return(post.Post{}, errors.New("vote error"))

		handler.VotePost(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var res map[string]string
		err := json.NewDecoder(w.Body).Decode(&res)
		if err != nil {
			return
		}
		assert.Equal(t, "vote error", res["message"])
	})

	t.Run("json encoding error on failure", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/upvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "upvote"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		mockRepo.EXPECT().VotePost("abc", "123", 1).Return(post.Post{}, errors.New("vote error"))

		handler.VotePost(&brokenWrite{}, req)
	})

	t.Run("json encoding error on success", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/posts/abc/votes/upvote", nil)
		req = mux.SetURLVars(req, map[string]string{"POST_ID": "abc", "VOTE_TYPE": "upvote"})
		req = req.WithContext(context.WithValue(req.Context(), ctxKey, claims))

		mockRepo.EXPECT().VotePost("abc", "123", 1).Return(post.Post{Title: "voted"}, nil)

		handler.VotePost(&brokenWrite{}, req)
	})
}
