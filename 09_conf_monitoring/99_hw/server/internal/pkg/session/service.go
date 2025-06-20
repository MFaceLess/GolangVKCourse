package session

import (
	"errors"
	"fmt"
	"net/http"
	"server/internal/pkg/domain"
	"time"

	"server/internal/metrics"

	log "github.com/sirupsen/logrus"
)

type service struct{}

func NewService() domain.SessionService {
	return service{}
}

func (s service) CheckSession(headers http.Header) (domain.Session, error) {
	const endpoint = "http://vk-golang.ru:17000/int/CheckSession"
	logEntry := log.WithFields(log.Fields{
		"service": "session",
		"method":  "CheckSession",
		"url":     endpoint,
	})

	logEntry.Debug("Starting session check")

	req, err := http.NewRequest(http.MethodGet, "http://vk-golang.ru:17000/int/CheckSession", nil)
	if err != nil {
		logEntry.WithError(err).Error("Failed to create request")
		return domain.Session{}, err
	}

	req.Header = headers

	startTime := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:17000", "500").Inc()
		metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:17000").Observe(time.Since(startTime).Seconds())

		logEntry.WithError(err).Error("Request failed")
		return domain.Session{}, err
	}

	metrics.ExternalServiceStatus.WithLabelValues("vk-golang.ru:17000", fmt.Sprintf("%d", resp.StatusCode)).Inc()
	metrics.ExternalServiceTimings.WithLabelValues("vk-golang.ru:17000").Observe(time.Since(startTime).Seconds())

	logEntry.WithFields(log.Fields{
		"status_code": resp.StatusCode,
		"duration":    time.Since(startTime),
	}).Debug("Received response")

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		logEntry.Error("Internal server error from session service")
		return domain.Session{}, errors.New("failed to request check session")
	case http.StatusOK:
		logEntry.Debug("Session check successful")
		return domain.Session{}, nil
	default:
		logEntry.WithField("status_code", resp.StatusCode).Warn("No valid session found")
		return domain.Session{}, domain.ErrNoSession
	}
}
