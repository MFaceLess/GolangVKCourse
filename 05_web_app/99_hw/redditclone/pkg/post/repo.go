package post

import (
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"
)

var (
	ErrPostNotFound   = errors.New("post not found")
	ErrCantGenerateID = errors.New("generate id error")
)

const (
	NoneCriterion     = ""
	CategoryCriterion = "category"
	AuthorCriterion   = "author"
)

type CriteriaData struct {
	Name string
	Data string
}

type PostMemoryRepository struct {
	data map[string]*Post
	*sync.RWMutex
}

func NewMemoryRepo() *PostMemoryRepository {
	return &PostMemoryRepository{
		data:    make(map[string]*Post, 10),
		RWMutex: &sync.RWMutex{},
	}
}

func FindItemsOnCriterion(repo *PostMemoryRepository, criterion *CriteriaData) []*Post {
	var items []*Post

	repo.RLock()
	defer repo.RUnlock()

	for _, post := range repo.data {
		switch criterion.Name {
		case NoneCriterion:
			items = append(items, post)
		case CategoryCriterion:
			if post.Category == criterion.Data {
				items = append(items, post)
			}
		case AuthorCriterion:
			if post.Author.Username == criterion.Data {
				items = append(items, post)
			}
		}
	}

	slices.SortFunc(items, func(a, b *Post) int {
		return b.Score - a.Score
	})

	return items
}

func (repo *PostMemoryRepository) GetAll() []*Post {
	criteriaData := &CriteriaData{Name: NoneCriterion}
	return FindItemsOnCriterion(repo, criteriaData)
}

func (repo *PostMemoryRepository) GetByID(id string) *Post {
	repo.RLock()
	defer repo.RUnlock()
	return repo.data[id]
}

func (repo *PostMemoryRepository) Add(postData DataPost, login string, userID string) *Post {
	randID := make([]byte, 16)
	if _, err := rand.Read(randID); err != nil {
		return nil
	}

	post := &Post{
		Score: 1,
		Views: 0,
		Type:  postData.Type,
		Title: postData.Title,
		Author: Author{
			Username: login,
			ID:       userID,
		},
		Category: postData.Category,
		Votes: []Vote{
			{
				User: userID,
				Vote: 1,
			},
		},
		Comments:         []Comment{},
		Created:          time.Now().UTC().Format(time.RFC3339Nano),
		UpvotePercentage: 100,
		ID:               fmt.Sprintf("%x", randID),
		URL:              postData.URL,
		Text:             postData.Text,
	}

	repo.Lock()
	repo.data[post.ID] = post
	repo.Unlock()

	return post
}

func (repo *PostMemoryRepository) GetByCategory(category string) []*Post {
	criteriaData := &CriteriaData{Name: CategoryCriterion, Data: category}
	return FindItemsOnCriterion(repo, criteriaData)
}

func (repo *PostMemoryRepository) GetPostsByUsername(username string) []*Post {
	CriteriaData := &CriteriaData{Name: AuthorCriterion, Data: username}
	return FindItemsOnCriterion(repo, CriteriaData)
}

func (repo *PostMemoryRepository) DeletePostByID(postID, username string) error {
	repo.Lock()
	defer repo.Unlock()

	post, ok := repo.data[postID]
	if !ok {
		return ErrSourceNotFound
	}

	if post.Author.Username != username {
		return ErrAccessDenied
	}

	delete(repo.data, postID)

	return nil
}

func (repo *PostMemoryRepository) AddComment(postID, body, username, userID string) (*Post, error) {
	randID := make([]byte, 16)
	if _, err := rand.Read(randID); err != nil {
		return nil, ErrCantGenerateID
	}

	comment := Comment{
		Created: time.Now().UTC().Format(time.RFC3339Nano),
		Author:  Author{Username: username, ID: userID},
		Body:    body,
		ID:      fmt.Sprintf("%x", randID),
	}

	repo.Lock()
	defer repo.Unlock()

	post, ok := repo.data[postID]
	if !ok {
		return nil, ErrPostNotFound
	}

	post.Comments = append(post.Comments, comment)

	return post, nil
}

func (repo *PostMemoryRepository) DeleteComment(postID, commentID, username string) (*Post, error) {

	repo.RLock()
	post, ok := repo.data[postID]
	repo.RUnlock()

	if !ok {
		return nil, ErrSourceNotFound
	}
	var indexToRemove int
	ok = false
	for i, comment := range post.Comments {
		if comment.ID != commentID {
			continue
		}

		if comment.Author.Username != username {
			return nil, ErrAccessDenied
		}
		indexToRemove = i
		ok = true
		break

	}
	if !ok {
		return nil, ErrSourceNotFound
	}

	repo.Lock()
	defer repo.Unlock()

	repo.data[postID].Comments = append(repo.data[postID].Comments[:indexToRemove], repo.data[postID].Comments[indexToRemove+1:]...)

	return repo.data[postID], nil
}

func (repo *PostMemoryRepository) VotePost(postID, userID string, voteDirection int) (*Post, error) {
	repo.RLock()
	post, ok := repo.data[postID]
	repo.RUnlock()
	if !ok {
		return nil, ErrSourceNotFound
	}

	var (
		existingIndex = -1
		existingVote  = VoteNone
	)

	for i, v := range post.Votes {
		if v.User == userID {
			existingIndex = i
			existingVote = v.Vote
			break
		}
	}

	repo.Lock()
	defer repo.Unlock()

	switch {
	case existingIndex == -1 && voteDirection == VoteNone:
		return post, nil

	case existingIndex == -1:
		repo.data[postID].Votes = append(repo.data[postID].Votes, Vote{
			User: userID,
			Vote: voteDirection,
		})
		repo.data[postID].Score += voteDirection

	case existingIndex != -1 && voteDirection == VoteNone:
		repo.data[postID].Score -= existingVote
		repo.data[postID].Votes = append(
			repo.data[postID].Votes[:existingIndex],
			repo.data[postID].Votes[existingIndex+1:]...,
		)

	case existingIndex != -1 && voteDirection == existingVote:
		return nil, ErrAlreadyVoted

	default:
		delta := voteDirection - existingVote
		repo.data[postID].Score += delta
		repo.data[postID].Votes[existingIndex].Vote = voteDirection

	}

	totalVotes := len(repo.data[postID].Votes)
	if totalVotes > 0 {
		ups := 0
		for _, v := range repo.data[postID].Votes {
			if v.Vote == VoteUp {
				ups++
			}
		}
		repo.data[postID].UpvotePercentage = ups * 100 / totalVotes
	} else {
		repo.data[postID].UpvotePercentage = 0
	}

	return repo.data[postID], nil
}
