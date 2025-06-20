package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"notestore/notestore/pkg/note"
	"notestore/notestore/pkg/response"
)

type NoteHandler struct {
	NoteRepo note.NoteRepo
}

func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.RespJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	note := h.NoteRepo.GetNote(idInt)

	if err = json.NewEncoder(w).Encode(note); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.RespJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	var noteData note.Note
	err := json.NewDecoder(r.Body).Decode(&noteData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.RespJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	note := h.NoteRepo.CreateNote(noteData.Text)

	if err = json.NewEncoder(w).Encode(note); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.RespJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.RespJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	var noteData note.Note
	err = json.NewDecoder(r.Body).Decode(&noteData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.RespJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	note := h.NoteRepo.UpdateNote(idInt, noteData.Text)

	if err = json.NewEncoder(w).Encode(note); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.RespJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	idInt, err := strconv.Atoi(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.RespJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.NoteRepo.DeleteNote(idInt)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response.RespJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	response.RespJSON(w, http.StatusNoContent, "note deleted")
}

func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	orderBy := queryParams.Get("order_by")

	if orderBy != note.OrderByID && orderBy != note.OrderByText && orderBy != note.OrderByCreated && orderBy != note.OrderByUpdated && orderBy != "" {
		w.WriteHeader(http.StatusBadRequest)
		response.RespJSON(w, http.StatusBadRequest, "order_by has an invalid data")
		return
	}

	notes := h.NoteRepo.GetNotes(orderBy)

	if err := json.NewEncoder(w).Encode(notes); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.RespJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
}
