package repository

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server/internal/metrics"
	"time"

	"server/internal/pkg/domain"

	log "github.com/sirupsen/logrus"
)

type repository struct{}

func NewRepository() domain.ThreadRepository {
	return repository{}
}

func (r repository) Create(thread domain.Thread) error {
	log.WithFields(log.Fields{
		"url":    "http://vk-golang.ru:15000/thread",
		"thread": fmt.Sprintf("%+v", thread),
	}).Debug("Creating thread")

	reqBody, err := json.Marshal(thread)
	if err != nil {
		log.WithError(err).Error("Failed to marshal thread")
		return fmt.Errorf("failed to marshal thread: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "http://vk-golang.ru:15000/thread", bytes.NewBuffer(reqBody))
	if err != nil {
		log.WithError(err).Error("Failed to create request")
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:15000", "500").Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:15000").Observe(time.Since(startTime).Seconds())
		log.WithError(err).Error("Request failed")
		return fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:15000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:15000").Observe(time.Since(startTime).Seconds())
		errMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		log.WithField("status_code", resp.StatusCode).Error(errMsg)
		return errors.New("failed to create thread remotely")
	}

	metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:15000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
	metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:15000").Observe(time.Since(startTime).Seconds())

	log.Info("Thread created successfully")
	return nil
}

func (r repository) Get(id string) (domain.Thread, error) {
	url := fmt.Sprintf("http://vk-golang.ru:15000/thread?id=%s", id)
	log.WithFields(log.Fields{
		"url": url,
		"id":  id,
	}).Debug("Fetching thread")

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://vk-golang.ru:15000/thread?id=%s", id), nil)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"id":    id,
		}).Error("Failed to create request")
		return domain.Thread{}, fmt.Errorf("failed to create request: %w", err)
	}

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:15000", "500").Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:15000").Observe(time.Since(startTime).Seconds())
		log.WithFields(log.Fields{
			"error": err,
			"id":    id,
		}).Error("Request failed")
		return domain.Thread{}, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:15000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:15000").Observe(time.Since(startTime).Seconds())
		errMsg := fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		log.WithFields(log.Fields{
			"status_code": resp.StatusCode,
			"id":          id,
		}).Error(errMsg)
		return domain.Thread{}, errors.New("failed to fetch thread remotely")
	}

	metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:15000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
	metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:15000").Observe(time.Since(startTime).Seconds())

	var thread domain.Thread
	err = json.NewDecoder(resp.Body).Decode(&thread)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"id":    id,
		}).Error("Failed to decode response")
		return domain.Thread{}, fmt.Errorf("failed to decode response: %w", err)
	}

	log.WithField("id", id).Info("Thread fetched successfully")
	return thread, nil
}
