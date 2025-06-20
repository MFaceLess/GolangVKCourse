package main

import (
	"fmt"
	"log"
	"net/http"
	"server/internal/api/middleware"
	"server/internal/pkg/comment/handler"
	commentrepo "server/internal/pkg/comment/repository"
	commentsvc "server/internal/pkg/comment/service"
	"server/internal/pkg/session"
	threadhttp "server/internal/pkg/thread/handler"
	threadrepo "server/internal/pkg/thread/repository"
	threadsvc "server/internal/pkg/thread/service"

	"github.com/labstack/echo"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	e := echo.New()
	sessionSvc := session.NewService()
	e.Use(middleware.AuthEchoMiddleware(sessionSvc))

	threadRepo := threadrepo.NewRepository()
	threadSvc := threadsvc.NewService(threadRepo)
	threadHandler := threadhttp.Handler{ThreadSvc: threadSvc}

	commentRepo := commentrepo.NewRepository()
	commentSvc := commentsvc.NewService(commentRepo, threadRepo)
	commentHandler := handler.Handler{CommentSvc: commentSvc}

	// e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("start server error: %v", err)
		}
	}()

	e.GET("/thread/:tid", threadHandler.GetThread)
	e.POST("/thread", threadHandler.CreateThread)
	e.POST("/thread/:tid/comment", commentHandler.Create)
	e.POST("/thread/:tid/comment/:cid/like", commentHandler.Like)

	fmt.Print(e.Start(":8000"))
}
