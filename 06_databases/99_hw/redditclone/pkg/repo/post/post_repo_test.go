package repo

import (
	"redditclone/pkg/post"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

const (
	userUsername     = "user"
	incorrectMessage = "incorrect"
)

func TestGetAll(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("return all posts", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		expectedPosts := []post.Post{
			{Title: "First", Score: 10},
			{Title: "Second", Score: 20},
		}

		docs := make([]bson.D, len(expectedPosts))
		for i, post := range expectedPosts {
			docs[i] = bson.D{
				{Key: "title", Value: post.Title},
				{Key: "score", Value: post.Score},
			}
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, docs...),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		posts, err := repo.GetAll()

		assert.NoError(t, err)
		assert.Len(t, posts, 2)

		assert.Equal(t, "First", posts[0].Title)
		assert.Equal(t, 10, posts[0].Score)

		assert.Equal(t, "Second", posts[1].Title)
		assert.Equal(t, 20, posts[1].Score)
	})

	mt.Run("error on Find", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: "find error",
			}),
		)

		posts, err := repo.GetAll()
		assert.Error(t, err)
		assert.Nil(t, posts)
	})

	mt.Run("error on cursor.All", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, bson.D{{Key: "title", Value: "Test"}, {Key: "score", Value: 10}}),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    2,
				Message: "cursor all error",
			}),
		)

		posts, err := repo.GetAll()
		assert.Nil(t, posts)
		assert.Error(t, err)
	})
}

func TestGetById(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		id := primitive.NewObjectID()
		expected := bson.D{
			{Key: "_id", Value: id},
			{Key: "Title", Value: "First"},
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, expected))

		post, err := repo.GetByID(id.Hex())
		assert.NoError(t, err)
		assert.Equal(t, id, post.ID)
		assert.Equal(t, "First", post.Title)
	})

	mt.Run("invalid hex ID format", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		_, err := repo.GetByID("invalid_hex")
		assert.Error(t, err)
	})

	mt.Run("post not found", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.posts", mtest.FirstBatch))

		_, err := repo.GetByID(primitive.NewObjectID().Hex())

		assert.ErrorIs(t, err, ErrPostNotFound)
	})

	mt.Run("internal FindOne error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    123,
			Message: "internal error",
		}))

		_, err := repo.GetByID(primitive.NewObjectID().Hex())

		assert.Error(t, err)
	})
}

func TestAdd(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postData := post.DataPost{
			Type:     "link",
			Title:    "Wee",
			Category: "music",
			URL:      "http://test.ru",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		post, err := repo.Add(postData, "temp", "1")

		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Equal(t, "Wee", post.Title)
		assert.Equal(t, "music", post.Category)
		assert.Equal(t, "http://test.ru", post.URL)
		assert.Equal(t, "link", post.Type)
		assert.Equal(t, 1, post.Score)
		assert.Equal(t, 1, post.Votes[0].Vote)
		assert.Equal(t, 100, post.UpvotePercentage)
	})

	mt.Run("insert fails with error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postData := post.DataPost{
			Type:     "text",
			Title:    "Fail case",
			Category: "test",
			Text:     "This should fail",
		}

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    11000,
			Message: "insert error",
		}))

		_, err := repo.Add(postData, "tester", "user456")

		assert.Error(t, err)
	})
}

func TestGetByCategory(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		expectedPosts := []post.Post{
			{Title: "First", Score: 10, Category: "music"},
			{Title: "Second", Score: 20, Category: "news"},
		}

		docs := make([]bson.D, len(expectedPosts))
		for i, post := range expectedPosts {
			docs[i] = bson.D{
				{Key: "title", Value: post.Title},
				{Key: "score", Value: post.Score},
				{Key: "category", Value: post.Category},
			}
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, docs[0]),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		posts, err := repo.GetByCategory("music")
		assert.NoError(t, err)
		assert.Len(t, posts, 1)
		assert.Equal(t, "First", posts[0].Title)
		assert.Equal(t, 10, posts[0].Score)
		assert.Equal(t, "music", posts[0].Category)
	})
}

func TestGetByUsername(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		expectedPosts := []post.Post{
			{Title: "First", Score: 10, Category: "music", Author: post.Author{ID: "123", Username: userUsername}},
			{Title: "Second", Score: 20, Category: "news", Author: post.Author{ID: "1234", Username: "user2"}},
		}

		docs := make([]bson.D, len(expectedPosts))
		for i, post := range expectedPosts {
			docs[i] = bson.D{
				{Key: "title", Value: post.Title},
				{Key: "score", Value: post.Score},
				{Key: "category", Value: post.Category},
			}
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, docs[1]),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		posts, err := repo.GetPostsByUsername("user2")
		assert.NoError(t, err)
		assert.Len(t, posts, 1)
		assert.Equal(t, "Second", posts[0].Title)
		assert.Equal(t, 20, posts[0].Score)
		assert.Equal(t, "news", posts[0].Category)
	})
}

func TestUndefinedCriteria(t *testing.T) {
	repo := NewMemoryRepo(nil)
	posts, err := findItemsOnCriterion(repo, &CriteriaData{Name: "Not Exists"})

	assert.Len(t, posts, 0)
	assert.ErrorIs(t, ErrUndefinedCriterion, err)
}

func TestDeletePostByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully deletes post", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		username := userUsername

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: postID},
			{Key: "author", Value: bson.D{{Key: "username", Value: username}}},
		}),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
			mtest.CreateSuccessResponse(),
		)

		err := repo.DeletePostByID(postID.Hex(), username)
		assert.NoError(t, err)
	})

	mt.Run("post not found", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		err := repo.DeletePostByID(postID.Hex(), userUsername)
		assert.ErrorIs(t, err, post.ErrSourceNotFound)
	})

	mt.Run("another error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(
			mtest.CommandError{
				Code:    12345,
				Name:    "SomeError",
				Message: "unexpected error",
			},
		))

		err := repo.DeletePostByID(postID.Hex(), userUsername)
		assert.Error(t, err)
		assert.NotEqual(t, post.ErrSourceNotFound, err)
	})

	mt.Run("incorred postID", func(mt *mtest.T) {
		repo := NewMemoryRepo(nil)

		postID := incorrectMessage

		err := repo.DeletePostByID(postID, userUsername)
		assert.Error(t, err)
	})

	mt.Run("user not author", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: postID},
			{Key: "author", Value: bson.D{{Key: "username", Value: "author"}}},
		}),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
			mtest.CreateSuccessResponse(),
		)

		err := repo.DeletePostByID(postID.Hex(), userUsername)
		assert.Error(t, err)
		assert.ErrorIs(t, err, post.ErrAccessDenied)
	})

	mt.Run("error delete one", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: postID},
			{Key: "author", Value: bson.D{{Key: "username", Value: "author"}}},
		}),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    123,
				Name:    "DeleteError",
				Message: "delete failed",
			}),
		)

		err := repo.DeletePostByID(postID.Hex(), "author")
		assert.Error(t, err)
	})
}

func TestAddComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully adds comment", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		userID := primitive.NewObjectID().Hex()
		username := "testuser"
		body := "Nice post!"

		updatedPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "comments", Value: bson.A{
				bson.D{
					{Key: "created", Value: time.Now().UTC().Format(time.RFC3339Nano)},
					{Key: "author", Value: bson.D{
						{Key: "username", Value: username},
						{Key: "id", Value: userID},
					}},
					{Key: "body", Value: body},
					{Key: "_id", Value: primitive.NewObjectID()},
				},
			}},
		}

		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "value", Value: updatedPost}})

		post, err := repo.AddComment(postID.Hex(), body, username, userID)
		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Len(t, post.Comments, 1)
		assert.Equal(t, body, post.Comments[0].Body)
		assert.Equal(t, username, post.Comments[0].Author.Username)
	})

	mt.Run("post not found", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch),
		)

		_, err := repo.AddComment(postID.Hex(), "hello", userUsername, "123")
		assert.ErrorIs(t, err, ErrPostNotFound)
	})

	mt.Run("another error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    123,
				Name:    "add comment error",
				Message: "add comment failed",
			}),
		)

		_, err := repo.AddComment(postID.Hex(), "hello", userUsername, "123")
		assert.Error(t, err)
	})

	mt.Run("incorrect postID", func(mt *mtest.T) {
		repo := NewMemoryRepo(nil)

		postID := incorrectMessage
		_, err := repo.AddComment(postID, "hello", userUsername, "123")

		assert.Error(t, err)
	})
}

func TestDeleteComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successfully deletes comment", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		commentID := primitive.NewObjectID()
		username := "testuser"

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "cursor", Value: bson.D{
				{Key: "firstBatch", Value: bson.A{
					bson.D{
						{Key: "_id", Value: postID},
						{Key: "comments", Value: bson.A{
							bson.D{
								{Key: "_id", Value: commentID},
								{Key: "body", Value: "test comment"},
								{Key: "author", Value: bson.D{
									{Key: "username", Value: username},
								}},
							},
						}},
					},
				}},
			}},
		})

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{
				{Key: "_id", Value: postID},
				{Key: "comments", Value: bson.A{}},
			}},
		})

		post, err := repo.DeleteComment(postID.Hex(), commentID.Hex(), username)
		assert.NoError(t, err)
		assert.NotNil(t, post)
		assert.Len(t, post.Comments, 0)
	})

	mt.Run("invalid postID", func(mt *mtest.T) {
		repo := NewMemoryRepo(nil)

		postID := incorrectMessage
		_, err := repo.DeleteComment(postID, primitive.NewObjectID().Hex(), userUsername)

		assert.Error(t, err)
	})

	mt.Run("invalid commentID", func(mt *mtest.T) {
		repo := NewMemoryRepo(nil)

		commentID := incorrectMessage
		_, err := repo.DeleteComment(primitive.NewObjectID().Hex(), commentID, userUsername)

		assert.Error(t, err)
	})

	mt.Run("error find one", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		commentID := primitive.NewObjectID()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(0, "db.posts", mtest.FirstBatch),
		)

		_, err := repo.DeleteComment(postID.Hex(), commentID.Hex(), userUsername)
		assert.ErrorIs(t, err, post.ErrSourceNotFound)
	})

	mt.Run("another error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		commentID := primitive.NewObjectID()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    123,
				Name:    "delete comment error",
				Message: "delete comment failed",
			}),
		)

		_, err := repo.DeleteComment(postID.Hex(), commentID.Hex(), userUsername)
		assert.Error(t, err)
	})

	mt.Run("user not author", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		commentID := primitive.NewObjectID()

		mockPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "comments", Value: bson.A{
				bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "body", Value: "test"},
					{Key: "author", Value: bson.D{
						{Key: "username", Value: "otheruser"},
					}},
				},
				bson.D{
					{Key: "_id", Value: commentID},
					{Key: "body", Value: "test"},
					{Key: "author", Value: bson.D{
						{Key: "username", Value: "otheruser"},
					}},
				},
			}},
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, mockPost),
		)

		_, err := repo.DeleteComment(postID.Hex(), commentID.Hex(), userUsername)
		assert.ErrorIs(t, err, post.ErrAccessDenied)
	})

	mt.Run("comment not found", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		commentID := primitive.NewObjectID()
		username := userUsername

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "cursor", Value: bson.D{
				{Key: "firstBatch", Value: bson.A{
					bson.D{
						{Key: "_id", Value: postID},
						{Key: "comments", Value: bson.A{
							bson.D{
								{Key: "_id", Value: commentID},
								{Key: "body", Value: "test comment"},
								{Key: "author", Value: bson.D{
									{Key: "username", Value: username},
								}},
							},
						}},
					},
				}},
			}},
		})

		_, err := repo.DeleteComment(postID.Hex(), primitive.NewObjectID().Hex(), username)
		assert.ErrorIs(t, err, post.ErrSourceNotFound)
	})

	mt.Run("error FindOneAndUpdate", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		postID := primitive.NewObjectID()
		commentID := primitive.NewObjectID()
		username := userUsername

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: postID},
			{Key: "comments", Value: bson.A{
				bson.D{
					{Key: "_id", Value: commentID},
					{Key: "body", Value: "test comment"},
					{Key: "author", Value: bson.D{
						{Key: "username", Value: username},
					}},
				},
			}},
		}))

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    11000,
			Name:    "FindOneAndUpdateError",
			Message: "some error",
		}))

		_, err := repo.DeleteComment(postID.Hex(), commentID.Hex(), username)
		assert.Error(t, err)
		assert.NotErrorIs(t, err, post.ErrSourceNotFound)
		assert.NotErrorIs(t, err, post.ErrAccessDenied)
	})
}

func TestVotePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	basePost := bson.D{
		{Key: "_id", Value: postID},
		{Key: "votes", Value: bson.A{}},
		{Key: "score", Value: 0},
	}

	mt.Run("invalid post id", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		_, err := repo.VotePost("invalid", userID.Hex(), post.VoteUp)
		assert.Error(t, err)
	})

	mt.Run("post not found", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.posts", mtest.FirstBatch))

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)
		assert.ErrorIs(t, err, post.ErrSourceNotFound)
	})

	mt.Run("another error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    11000,
			Name:    "FindOneAndUpdateError",
			Message: "some error",
		}))

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)
		assert.Error(t, err)
	})

	mt.Run("new vote", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		post, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)
		assert.NoError(t, err)
		assert.NotNil(t, post)
	})

	mt.Run("cancel existing vote", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		votedPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "votes", Value: bson.A{
				bson.D{{Key: "_id", Value: userID}, {Key: "vote", Value: post.VoteUp}},
			}},
			{Key: "score", Value: 1},
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		post, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteNone)
		assert.NoError(t, err)
		assert.NotNil(t, post)
	})

	mt.Run("same vote again", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)
		votedPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "votes", Value: bson.A{
				bson.D{{Key: "_id", Value: userID}, {Key: "vote", Value: post.VoteDown}},
			}},
			{Key: "score", Value: -1},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost))

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteDown)
		assert.ErrorIs(t, err, post.ErrAlreadyVoted)
	})

	mt.Run("change vote", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		votedPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "votes", Value: bson.A{
				bson.D{{Key: "_id", Value: userID}, {Key: "vote", Value: post.VoteDown}},
			}},
			{Key: "score", Value: -1},
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, votedPost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		post, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)
		assert.NoError(t, err)
		assert.NotNil(t, post)
	})

	mt.Run("no vote and VoteNone", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		testPost := bson.D{
			{Key: "_id", Value: postID},
			{Key: "votes", Value: bson.A{}},
			{Key: "score", Value: 0},
		}

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, testPost),
		)

		post, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteNone)

		assert.NoError(t, err)
		assert.NotNil(t, post)

		assert.Empty(t, post.Votes)
		assert.Equal(t, 0, post.Score)
	})

	mt.Run("error while finding post", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost))

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Code:    11000,
			Message: "Error updating post",
		}))

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)

		assert.Error(t, err)
	})

	mt.Run("error when finding post after update", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost))
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.posts", mtest.FirstBatch))

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)

		assert.Error(t, err)
	})

	mt.Run("error when finding post after update", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),
		)

		mt.AddMockResponses(mtest.CreateCommandErrorResponse(mtest.CommandError{
			Code:    11000,
			Name:    "UpdateError",
			Message: "Failed to update document",
		}))

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)

		assert.Error(t, err)
	})

	mt.Run("FindOne error", func(mt *mtest.T) {
		repo := NewMemoryRepo(mt.Coll)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),

			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch, basePost),
			mtest.CreateCursorResponse(0, "db.posts", mtest.NextBatch),

			mtest.CreateSuccessResponse(),
			mtest.CreateCursorResponse(1, "db.posts", mtest.FirstBatch),
		)

		_, err := repo.VotePost(postID.Hex(), userID.Hex(), post.VoteUp)

		assert.Error(t, err)
	})
}
