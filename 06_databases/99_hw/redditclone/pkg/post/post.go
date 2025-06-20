package post

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	VoteUp   = 1
	VoteDown = -1
	VoteNone = 0
)

var (
	ErrAccessDenied   = errors.New("у вас нет прав на данное действие")
	ErrSourceNotFound = errors.New("post not found")
	ErrAlreadyVoted   = errors.New("вы уже сделали такой голос")
)

type Author struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
}

type Vote struct {
	User string `json:"user" bson:"_id"`
	Vote int    `json:"vote" bson:"vote"`
}

type Comment struct {
	Created string             `json:"created" bson:"created"`
	Author  Author             `json:"author" bson:"author"`
	Body    string             `json:"body" bson:"body"`
	ID      primitive.ObjectID `json:"id" bson:"_id"`
}

type DataComment struct {
	Comment string `json:"comment"`
}

type Post struct {
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type"`
	Title            string             `json:"title" bson:"title"`
	URL              string             `json:"url,omitempty" bson:"url,omitempty"`
	Author           Author             `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category"`
	Text             string             `json:"text,omitempty" bson:"text,omitempty"`
	Votes            []Vote             `json:"votes" bson:"votes"`
	Comments         []Comment          `json:"comments" bson:"comments"`
	Created          string             `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               primitive.ObjectID `json:"id" bson:"_id"`
}

type DataPost struct {
	Category string `json:"category"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	Text     string `json:"text,omitempty"`
}