package handler

import (
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"

	"server/internal/metrics"
	"server/internal/pkg/domain"
)

type Handler struct {
	ThreadSvc domain.ThreadService
}

func (h Handler) GetThread(ctx echo.Context) error {
	startTime := time.Now()

	tid := ctx.Param("tid")

	requestID, ok := ctx.Get("requestID").(string)
	if !ok {
		requestID = ""
	}

	logEntry := log.WithFields(log.Fields{
		"handler":    "GetThread",
		"request_id": requestID,
		"thread_id":  tid,
	})

	logEntry.Debug("Start processing request")

	t, err := h.ThreadSvc.Get(tid)
	if err != nil {
		metrics.HitsTotal.WithLabelValues("GetThread", "500").Inc()
		metrics.HitTimings.WithLabelValues("GetThread").Observe(time.Since(startTime).Seconds())

		logEntry.WithError(err).Error("Failed to get thread")
		return err
	}

	metrics.HitsTotal.WithLabelValues("GetThread", "200").Inc()
	metrics.HitTimings.WithLabelValues("GetThread").Observe(time.Since(startTime).Seconds())

	logEntry.Info("Successfully processed request")
	return ctx.JSON(200, t)
}

func (h Handler) CreateThread(ctx echo.Context) error {
	startTime := time.Now()

	var thread domain.Thread

	requestID, ok := ctx.Get("requestID").(string)
	if !ok {
		requestID = ""
	}

	logEntry := log.WithFields(log.Fields{
		"handler":    "CreateThread",
		"request_id": requestID,
	})

	logEntry.Debug("Start processing request")

	err := ctx.Bind(&thread)
	if err != nil {
		metrics.HitsTotal.WithLabelValues("CreateThread", "500").Inc()
		metrics.HitTimings.WithLabelValues("CreateThread").Observe(time.Since(startTime).Seconds())

		logEntry.WithError(err).Error("Failed to bind request body")
		return err
	}

	err = h.ThreadSvc.Create(thread)
	if err != nil {
		metrics.HitsTotal.WithLabelValues("CreateThread", "500").Inc()
		metrics.HitTimings.WithLabelValues("CreateThread").Observe(time.Since(startTime).Seconds())

		logEntry.WithFields(log.Fields{
			"thread": thread,
		}).WithError(err).Error("Failed to create thread")
		return err
	}

	metrics.HitsTotal.WithLabelValues("CreateThread", "200").Inc()
	metrics.HitTimings.WithLabelValues("CreateThread").Observe(time.Since(startTime).Seconds())

	logEntry.Info("Thread created successfully")
	return ctx.NoContent(200)
}
