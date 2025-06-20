package handler

import (
	"server/internal/metrics"
	"server/internal/pkg/domain"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type Handler struct {
	CommentSvc domain.CommentService
}

func NewHandler(commentSvc domain.CommentService) *Handler {
	return &Handler{
		CommentSvc: commentSvc,
	}
}

func (h Handler) Create(ctx echo.Context) error {
	startTime := time.Now()
	defer func() {
		metrics.HitTimings.WithLabelValues("Create").Observe(time.Since(startTime).Seconds())
	}()

	requestID, ok := ctx.Get("requestID").(string)
	if !ok {
		requestID = ""
	}

	var comment domain.Comment

	log.WithFields(log.Fields{
		"request_id": requestID,
	}).Info("start Create Comment")

	err := ctx.Bind(&comment)
	if err != nil {
		metrics.HitsTotal.WithLabelValues("Create", "500").Inc()

		log.WithFields(log.Fields{
			"request_id": requestID,
			"error":      err,
		}).Error("error bind Create Comment")
		return err
	}

	tid := ctx.Param("tid")

	log.WithFields(log.Fields{
		"request_id": requestID,
		"thread_id":  tid,
	}).Info("Create Comment")
	metrics.HitsTotal.WithLabelValues("Create", "200").Inc()

	return h.CommentSvc.Create(tid, comment)
}

func (h Handler) Like(ctx echo.Context) error {
	startTime := time.Now()

	requestID, ok := ctx.Get("requestID").(string)
	if !ok {
		requestID = ""
	}

	tid := ctx.Param("tid")
	cid := ctx.Param("cid")

	log.WithFields(log.Fields{
		"request_id": requestID,
		"thread_id":  tid,
		"comment_id": cid,
	}).Info("start Like Comment")

	metrics.HitsTotal.WithLabelValues("Like", "200").Inc()
	metrics.HitTimings.WithLabelValues("Like").Observe(time.Since(startTime).Seconds())

	return h.CommentSvc.Like(tid, cid)
}
