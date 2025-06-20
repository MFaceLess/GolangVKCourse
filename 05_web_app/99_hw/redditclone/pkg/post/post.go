package post

import "errors"

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
	Username string `json:"username"`
	ID       string `json:"id"`
}

type Vote struct {
	User string `json:"user"`
	Vote int    `json:"vote"`
}

type Comment struct {
	Created string `json:"created"`
	Author  Author `json:"author"`
	Body    string `json:"body"`
	ID      string `json:"id"`
}

type DataComment struct {
	Comment string `json:"comment"`
}

type Post struct {
	Score            int       `json:"score"`
	Views            int       `json:"views"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	URL              string    `json:"url,omitempty"`
	Author           Author    `json:"author"`
	Category         string    `json:"category"`
	Text             string    `json:"text,omitempty"`
	Votes            []Vote    `json:"votes"`
	Comments         []Comment `json:"comments"`
	Created          string    `json:"created"`
	UpvotePercentage int       `json:"upvotePercentage"`
	ID               string    `json:"id"`
}

type DataPost struct {
	Category string `json:"category"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	URL      string `json:"url,omitempty"`
	Text     string `json:"text,omitempty"`
}

type PostRepo interface {
	GetAll() []*Post
	GetByID(id string) *Post
	Add(post DataPost, login string, userID string) *Post
	GetByCategory(category string) []*Post
	GetPostsByUsername(username string) []*Post
	DeletePostByID(postID, username string) error
	AddComment(postID, body, username, userID string) (*Post, error)
	DeleteComment(postID, commentID, username string) (*Post, error)
	VotePost(postID, userID string, voteDirection int) (*Post, error)
}
