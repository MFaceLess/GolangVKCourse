package repo

import (
	"context"
	"errors"
	"redditclone/pkg/post"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPostNotFound       = errors.New("post not found")
	ErrCantGenerateID     = errors.New("generate id error")
	ErrUndefinedCriterion = errors.New("invalid criterion")
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

func NewMemoryRepo(posts *mongo.Collection) *PostMemoryRepository {
	return &PostMemoryRepository{
		posts: posts,
	}
}

func findItemsOnCriterion(repo *PostMemoryRepository, criterion *CriteriaData) ([]post.Post, error) {
	var filter bson.M

	switch criterion.Name {
	case NoneCriterion:
		filter = bson.M{}
	case CategoryCriterion:
		filter = bson.M{"category": criterion.Data}
	case AuthorCriterion:
		filter = bson.M{"author.username": criterion.Data}
	default:
		return nil, ErrUndefinedCriterion

	}

	opts := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})

	cursor, err := repo.posts.Find(context.Background(), filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var posts []post.Post
	if err := cursor.All(context.Background(), &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (repo *PostMemoryRepository) GetAll() ([]post.Post, error) {
	return findItemsOnCriterion(repo, &CriteriaData{Name: NoneCriterion})
}

func (repo *PostMemoryRepository) GetByID(id string) (post.Post, error) {
	idObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return post.Post{}, err
	}

	var p post.Post
	err = repo.posts.FindOne(context.Background(), bson.M{"_id": idObj}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return post.Post{}, ErrPostNotFound
		}
		return post.Post{}, err
	}
	return p, nil
}

func (repo *PostMemoryRepository) Add(postData post.DataPost, login string, userID string) (post.Post, error) {
	p := post.Post{
		Score: 1,
		Views: 0,
		Type:  postData.Type,
		Title: postData.Title,
		Author: post.Author{
			Username: login,
			ID:       userID,
		},
		Category: postData.Category,
		Votes: []post.Vote{
			{
				User: userID,
				Vote: 1,
			},
		},
		Comments:         []post.Comment{},
		Created:          time.Now().UTC().Format(time.RFC3339Nano),
		UpvotePercentage: 100,
		ID:               primitive.NewObjectID(),
		URL:              postData.URL,
		Text:             postData.Text,
	}

	_, err := repo.posts.InsertOne(context.Background(), p)
	if err != nil {
		return post.Post{}, err
	}

	return p, nil
}

func (repo *PostMemoryRepository) GetByCategory(category string) ([]post.Post, error) {
	criteriaData := &CriteriaData{Name: CategoryCriterion, Data: category}
	return findItemsOnCriterion(repo, criteriaData)
}

func (repo *PostMemoryRepository) GetPostsByUsername(username string) ([]post.Post, error) {
	CriteriaData := &CriteriaData{Name: AuthorCriterion, Data: username}
	return findItemsOnCriterion(repo, CriteriaData)
}

func (repo *PostMemoryRepository) DeletePostByID(postID string, username string) error {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return err
	}

	var p post.Post
	err = repo.posts.FindOne(context.Background(), bson.M{"_id": postObjID}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return post.ErrSourceNotFound
		}
		return err
	}

	if p.Author.Username != username {
		return post.ErrAccessDenied
	}

	_, err = repo.posts.DeleteOne(context.Background(), bson.M{"_id": postObjID})
	if err != nil {
		return err
	}

	return nil
}

func (repo *PostMemoryRepository) AddComment(postID string, body, username string, userID string) (post.Post, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return post.Post{}, err
	}

	comment := post.Comment{
		Created: time.Now().UTC().Format(time.RFC3339Nano),
		Author:  post.Author{Username: username, ID: userID},
		Body:    body,
		ID:      primitive.NewObjectID(),
	}

	update := bson.M{
		"$push": bson.M{
			"comments": comment,
		},
	}

	var updatedPost post.Post
	err = repo.posts.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": postObjID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedPost)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return post.Post{}, ErrPostNotFound
		}
		return post.Post{}, err
	}

	return updatedPost, err
}

func (repo *PostMemoryRepository) DeleteComment(postID, commentID, username string) (post.Post, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return post.Post{}, err
	}

	commentObjID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		return post.Post{}, err
	}

	var p post.Post
	err = repo.posts.FindOne(context.Background(), bson.M{"_id": postObjID}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return post.Post{}, post.ErrSourceNotFound
		}
		return post.Post{}, err
	}

	var commentFound bool
	for _, comment := range p.Comments {
		if comment.ID != commentObjID {
			continue
		}

		if comment.Author.Username != username {
			return post.Post{}, post.ErrAccessDenied
		}
		commentFound = true
		break

	}

	if !commentFound {
		return post.Post{}, post.ErrSourceNotFound
	}

	update := bson.M{
		"$pull": bson.M{
			"comments": bson.M{"_id": commentObjID},
		},
	}

	var updatedPost post.Post
	err = repo.posts.FindOneAndUpdate(
		context.Background(),
		bson.M{"_id": postObjID},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedPost)

	if err != nil {
		return post.Post{}, err
	}

	return updatedPost, nil
}

func (repo *PostMemoryRepository) VotePost(postID, userID string, voteDirection int) (post.Post, error) {
	postObjID, err := primitive.ObjectIDFromHex(postID)
	if err != nil {
		return post.Post{}, err
	}

	ctx := context.Background()

	var p post.Post
	err = repo.posts.FindOne(ctx, bson.M{"_id": postObjID}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return post.Post{}, post.ErrSourceNotFound
		}
		return post.Post{}, err
	}

	var existingVote *post.Vote
	for _, v := range p.Votes {
		if v.User == userID {
			existingVote = &post.Vote{User: v.User, Vote: v.Vote}
			break
		}
	}

	var update bson.M
	var filter = bson.M{"_id": postObjID}

	switch {
	case existingVote == nil && voteDirection == post.VoteNone:
		return p, nil

	case existingVote == nil:
		update = bson.M{
			"$push": bson.M{"votes": bson.M{"_id": userID, "vote": voteDirection}},
			"$inc":  bson.M{"score": voteDirection},
		}
		_, err = repo.posts.UpdateOne(ctx, filter, update)

	case voteDirection == post.VoteNone:
		update = bson.M{
			"$pull": bson.M{"votes": bson.M{"_id": userID}},
			"$inc":  bson.M{"score": -existingVote.Vote},
		}
		_, err = repo.posts.UpdateOne(ctx, filter, update)

	case voteDirection == existingVote.Vote:
		return post.Post{}, post.ErrAlreadyVoted

	default:
		delta := voteDirection - existingVote.Vote
		update = bson.M{
			"$set": bson.M{"votes.$[elem].vote": voteDirection},
			"$inc": bson.M{"score": delta},
		}
		opts := options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{bson.M{"elem._id": userID}},
		})
		_, err = repo.posts.UpdateOne(ctx, filter, update, opts)
	}

	if err != nil {
		return post.Post{}, err
	}

	err = repo.posts.FindOne(ctx, bson.M{"_id": postObjID}).Decode(&p)
	if err != nil {
		return post.Post{}, err
	}

	upvotePercentage := calculateUpvotePercentage(p.Votes)

	_, err = repo.posts.UpdateOne(ctx,
		bson.M{"_id": postObjID},
		bson.M{"$set": bson.M{
			"upvotePercentage": upvotePercentage,
		}},
	)
	if err != nil {
		return post.Post{}, err
	}

	err = repo.posts.FindOne(ctx, bson.M{"_id": postObjID}).Decode(&p)
	if err != nil {
		return post.Post{}, err
	}

	return p, nil
}

func calculateUpvotePercentage(votes []post.Vote) int {
	total := len(votes)
	if total == 0 {
		return 0
	}

	upvotes := 0
	for _, v := range votes {
		if v.Vote == post.VoteUp {
			upvotes++
		}
	}

	return upvotes * 100 / total
}
