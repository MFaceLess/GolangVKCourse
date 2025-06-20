package note

import "time"

type Note struct {
	ID        int       `json:"id,omitempty"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type NoteRepo interface {
	GetNote(id int) *Note
	CreateNote(text string) *Note
	UpdateNote(id int, text string) *Note
	DeleteNote(id int) error
	GetNotes(parameter string) []*Note
}
