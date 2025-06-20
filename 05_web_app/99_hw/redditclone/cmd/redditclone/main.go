package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	"redditclone/pkg/post"
	"redditclone/pkg/user"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
}

func main() {
	userRepo := user.NewMemoryRepo()
	postRepo := post.NewMemoryRepo()

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
	}

	postHandler := &handlers.PostHandler{
		PostRepo: postRepo,
	}

	r := mux.NewRouter()

	s := http.StripPrefix("/static/", http.FileServer(http.Dir("../../static")))
	r.PathPrefix("/static/").Handler(s)
	r.Handle("/", http.FileServer(http.Dir("../../static/html")))
	r.Handle("/manifest.json", http.FileServer(http.Dir("../../static/")))

	r.HandleFunc("/api/login", userHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/api/register", userHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/api/posts/", postHandler.GetAllPosts).Methods(http.MethodGet)
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", postHandler.GetPostByCategory).Methods(http.MethodGet)
	r.HandleFunc("/api/post/{POST_ID}", postHandler.GetPostByID).Methods(http.MethodGet)
	r.HandleFunc("/api/user/{USER_LOGIN}", postHandler.GetPostsByUsername).Methods(http.MethodGet)

	protectedRouter := r.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.JWTMiddleWare)

	protectedRouter.HandleFunc("/posts", postHandler.AddPost).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/post/{POST_ID}", postHandler.DeletePost).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/post/{POST_ID}", postHandler.AddComment).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/post/{POST_ID}/{COMMENT_ID}", postHandler.DeleteComment).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/post/{POST_ID}/{VOTE_TYPE}", postHandler.VotePost).Methods(http.MethodGet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("starting server at :%s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		panic(err)
	}
}
