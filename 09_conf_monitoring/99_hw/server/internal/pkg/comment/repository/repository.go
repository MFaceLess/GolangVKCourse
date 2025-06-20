package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server/internal/pkg/domain"
	"time"

	"server/internal/metrics"

	log "github.com/sirupsen/logrus"
)

type loggingInterceptor struct {
	next domain.CommentRepository
}

func WithLogging(repo domain.CommentRepository) domain.CommentRepository {
	return &loggingInterceptor{
		next: repo,
	}
}

func (i *loggingInterceptor) Create(comment domain.Comment) error {
	log.WithFields(log.Fields{
		"method":  "Create",
		"comment": fmt.Sprintf("%+v", comment),
	}).Debug("Input parameters")

	startTime := time.Now()
	err := i.next.Create(comment)
	duration := time.Since(startTime)

	logFields := log.Fields{
		"duration": duration,
	}

	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Error("Create failed")
		return err
	}

	log.WithFields(logFields).Info("Create completed successfully")
	return nil
}

func (i *loggingInterceptor) Like(commentID string) error {
	log.WithFields(log.Fields{
		"method":     "Like",
		"comment_id": commentID,
	}).Debug("Input parameters")

	startTime := time.Now()
	err := i.next.Like(commentID)
	duration := time.Since(startTime)

	logFields := log.Fields{
		"duration": duration,
	}

	if err != nil {
		logFields["error"] = err
		log.WithFields(logFields).Error("Like failed")
		return err
	}

	log.WithFields(logFields).Info("Like completed successfully")
	return nil
}

type repository struct{}

func NewRepository() domain.CommentRepository {
	repo := repository{}
	return WithLogging(repo)
}

func (r repository) Create(comment domain.Comment) error {
	reqBody, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("failed to marshal comment: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "http://vk-golang.ru:16000/comment", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to construct request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:16000", "500").Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:16000").Observe(time.Since(startTime).Seconds())

		return fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:16000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:16000").Observe(time.Since(startTime).Seconds())

		return errors.New("failed to create comment remotely")
	}

	metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:16000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
	metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:16000").Observe(time.Since(startTime).Seconds())

	return nil
}

func (r repository) Like(commentID string) error {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://vk-golang.ru:16000/comment/like?cid=%s", commentID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to construct request: %w", err)
	}

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:16000", "500").Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:16000").Observe(time.Since(startTime).Seconds())

		return fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:16000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:16000").Observe(time.Since(startTime).Seconds())

		return errors.New("failed to like comment remotely")
	}

	metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:16000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
	metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:16000").Observe(time.Since(startTime).Seconds())

	return nil
}
