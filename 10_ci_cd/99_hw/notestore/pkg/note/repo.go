package note

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	OrderByID      = "id"
	OrderByText    = "text"
	OrderByCreated = "created_at"
	OrderByUpdated = "updated_at"
)

type NoteMemoryRepository struct {
	data   map[int]*Note
	lastID int
	mu     *sync.RWMutex
}

func NewMemoryRepo() *NoteMemoryRepository {
	return &NoteMemoryRepository{
		data: make(map[int]*Note, 10),
		mu:   &sync.RWMutex{},
	}
}

func (repo *NoteMemoryRepository) GetNote(id int) *Note {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	if note, ok := repo.data[id]; ok {
		return note
	}
	return nil
}

func (repo *NoteMemoryRepository) CreateNote(text string) *Note {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if repo.lastID == 0 {
		repo.lastID++
	}

	insertedNote := &Note{
		ID:        repo.lastID,
		Text:      text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.data[repo.lastID] = insertedNote
	repo.lastID += 1

	return insertedNote
}

func (repo *NoteMemoryRepository) UpdateNote(id int, text string) *Note {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	_, ok := repo.data[id]
	if !ok {
		return nil
	}

	repo.data[id].Text = text
	repo.data[id].UpdatedAt = time.Now()

	return repo.data[id]
}

func (repo *NoteMemoryRepository) DeleteNote(id int) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	_, ok := repo.data[id]
	if !ok {
		return errors.New("Note with this ID doesn't exist")
	}

	delete(repo.data, id)

	return nil
}

func (repo *NoteMemoryRepository) GetNotes(parameter string) []*Note {
	notes := make([]*Note, 0, len(repo.data))

	repo.mu.RLock()
	for _, note := range repo.data {
		notes = append(notes, note)
	}
	repo.mu.RUnlock()

	switch parameter {
	case OrderByID:
		slices.SortFunc(notes, func(a, b *Note) int {
			return a.ID - b.ID
		})
	case OrderByText:
		slices.SortFunc(notes, func(a, b *Note) int {
			return strings.Compare(a.Text, b.Text)
		})
	case OrderByCreated:
		slices.SortFunc(notes, func(a, b *Note) int {
			return a.CreatedAt.Compare(b.CreatedAt)
		})
	case OrderByUpdated:
		slices.SortFunc(notes, func(a, b *Note) int {
			return a.UpdatedAt.Compare(b.UpdatedAt)
		})

	}

	return notes
}
