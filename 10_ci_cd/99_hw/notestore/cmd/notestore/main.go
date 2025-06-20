package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"notestore/notestore/pkg/handlers"
	"notestore/notestore/pkg/middleware"
	"notestore/notestore/pkg/note"
)

const (
	PORT = "PORT"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found")
	}
}

func main() {
	noteRepo := note.NewMemoryRepo()
	noteHandler := &handlers.NoteHandler{
		NoteRepo: noteRepo,
	}

	router := createRouter(noteHandler)

	port := os.Getenv(PORT)
	if port == "" {
		port = "8080"
	}

	log.Printf("starting server at %s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("start server error: %v", err)
	}
}

func createRouter(noteHandler *handlers.NoteHandler) *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.Panic)

	r.HandleFunc("/note/{id}", noteHandler.GetNote).Methods(http.MethodGet)
	r.HandleFunc("/note", noteHandler.CreateNote).Methods(http.MethodPost)
	r.HandleFunc("/note/{id}", noteHandler.UpdateNote).Methods(http.MethodPut)
	r.HandleFunc("/note/{id}", noteHandler.DeleteNote).Methods(http.MethodDelete)
	r.HandleFunc("/note", noteHandler.GetNotes).Methods(http.MethodGet)

	return r
}
