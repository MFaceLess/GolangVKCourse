package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"redditclone/pkg/handlers"
	"redditclone/pkg/middleware"
	repoPost "redditclone/pkg/repo/post"
	repoUser "redditclone/pkg/repo/user"
	"redditclone/pkg/session"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
}

func main() {
	db, err := getMySQLDriver()
	if err != nil {
		log.Fatalf("error configure and start mysql db: %v", err)
	}

	sess, err := getMongoDBSession()
	if err != nil {
		log.Fatalf("error configure and start mongo db session: %v", err)
	}

	collection := sess.Database(os.Getenv("MONGO_INITDB_DATABASE")).Collection("posts")

	sessRepo := session.NewSessionMySQLRepo(db)
	jwtGen := session.Session{DBDriver: sessRepo}

	userRepo := repoUser.NewMemoryRepo(db, &jwtGen)
	postRepo := repoPost.NewMemoryRepo(collection)

	userHandler := &handlers.UserHandler{
		UserRepo: userRepo,
	}

	postHandler := &handlers.PostHandler{
		PostRepo: postRepo,
	}

	router := createRouter(db, userHandler, postHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server at :%s\n", port)
	if err = http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("start server error: %v", err)
	}
}

func getMySQLDriver() (*sql.DB, error) {
	dsn := os.Getenv("DB_DSN")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func getMongoDBSession() (*mongo.Client, error) {
	ctx := context.Background()
	sess, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_CONNECT")))
	if err != nil {
		return nil, err
	}

	err = sess.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func createRouter(db *sql.DB, userHandler *handlers.UserHandler, postHandler *handlers.PostHandler) *mux.Router {
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
	protectedRouter.Use(middleware.JWTMiddleWare(db))

	protectedRouter.HandleFunc("/posts", postHandler.AddPost).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/post/{POST_ID}", postHandler.DeletePost).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/post/{POST_ID}", postHandler.AddComment).Methods(http.MethodPost)
	protectedRouter.HandleFunc("/post/{POST_ID}/{COMMENT_ID}", postHandler.DeleteComment).Methods(http.MethodDelete)
	protectedRouter.HandleFunc("/post/{POST_ID}/{VOTE_TYPE}", postHandler.VotePost).Methods(http.MethodGet)

	return r
}
