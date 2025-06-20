package service

import (
	"server/internal/pkg/domain"

	log "github.com/sirupsen/logrus"
)

type service struct {
	CommentRepo domain.CommentRepository
	ThreadRepo  domain.ThreadRepository
}

func NewService(commentRepo domain.CommentRepository, threadRepo domain.ThreadRepository) domain.CommentService {
	return service{
		CommentRepo: commentRepo,
		ThreadRepo:  threadRepo,
	}
}

func (s service) Create(threadID string, comment domain.Comment) error {
	logEntry := log.WithFields(log.Fields{
		"service":    "comment",
		"method":     "Create",
		"thread_id":  threadID,
		"comment_id": comment.ID,
	})

	logEntry.Debug("Creating new comment")

	if err := s.checkThread(threadID); err != nil {
		logEntry.WithError(err).Error("Thread check failed")
		return err
	}

	logEntry.Info("Comment created successfully")
	return s.CommentRepo.Create(comment)
}

func (s service) Like(threadID string, commentID string) error {
	logEntry := log.WithFields(log.Fields{
		"service":    "comment",
		"method":     "Like",
		"thread_id":  threadID,
		"comment_id": commentID,
	})

	logEntry.Debug("Processing like for comment")

	if err := s.checkThread(threadID); err != nil {
		logEntry.WithError(err).Error("Thread check failed")
		return err
	}

	logEntry.Info("Comment liked successfully")
	return s.CommentRepo.Like(commentID)
}

func (s service) checkThread(threadID string) error {
	_, err := s.ThreadRepo.Get(threadID)
	return err
}
