package repo

import (
	"go.mongodb.org/mongo-driver/mongo"

	p "redditclone/pkg/post"
)

type PostRepo interface {
	GetAll() ([]p.Post, error)
	GetByID(id string) (p.Post, error)
	Add(postData p.DataPost, login string, userID string) (p.Post, error)
	GetByCategory(category string) ([]p.Post, error)
	GetPostsByUsername(username string) ([]p.Post, error)
	DeletePostByID(postID, username string) error
	AddComment(postID, body, username, userID string) (p.Post, error)
	DeleteComment(postID, commentID, username string) (p.Post, error)
	VotePost(postID, userID string, voteDirection int) (p.Post, error)
}

type PostMemoryRepository struct {
	posts *mongo.Collection
}
