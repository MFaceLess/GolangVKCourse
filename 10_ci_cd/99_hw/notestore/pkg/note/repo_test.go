package note

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoteMemoryRepository(t *testing.T) {
	repo := NewMemoryRepo()

	t.Run("CreateNote", func(t *testing.T) {
		note := repo.CreateNote("Test Note")
		assert.NotNil(t, note)
		assert.Equal(t, note.Text, "Test Note")
		assert.Equal(t, note.ID, 1)
	})
}
