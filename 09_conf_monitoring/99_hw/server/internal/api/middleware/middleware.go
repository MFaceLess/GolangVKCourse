package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"

	"server/internal/pkg/domain"
)

func AuthEchoMiddleware(service domain.SessionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {
			_, err := service.CheckSession(context.Request().Header)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
					"path":  context.Path(),
				}).Error("CheckSession failed")
				return context.NoContent(401)
			}

			context.Set("requestID", uuid.New().String())

			return next(context)
		}
	}
}
