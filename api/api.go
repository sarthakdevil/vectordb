package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"vectordb/internal/storage"
)

var store = storage.NewVectorStore("ivf")

type responseMessage struct {
	Type   string         `json:"type"`
	Text   string         `json:"text"`
	Vector *storage.Vector `json:"vector,omitempty"`
}

type insertRequest struct {
	Data []float64 `json:"data"`
	Text string    `json:"text"`
}

type searchRequest struct {
	Data   []float64 `json:"data"`
	Type   string    `json:"type"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, responseMessage{Type: "error", Text: msg})
}

func createindex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST allowed")
		return
	}

	indexType := r.URL.Query().Get("type")
	if indexType == "" {
		indexType = "ivf"
	}

	store = storage.NewVectorStore(indexType)
	writeJSON(w, http.StatusOK, responseMessage{
		Type: "success",
		Text: fmt.Sprintf("%s index created", indexType),
	})
}

func insertvectors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST allowed")
		return
	}

	var req insertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	data := make([]float32, len(req.Data))
	for i, v := range req.Data {
		data[i] = float32(v)
	}

	v := store.Add(data, req.Text)
	writeJSON(w, http.StatusCreated, responseMessage{
		Type: "success",
		Text: fmt.Sprintf("inserted vector %d", v.GetID()),
		Vector: &v,
	})
}

func Deletevectors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "only DELETE allowed")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if store.DeleteByID(id) {
		writeJSON(w, http.StatusOK, responseMessage{
			Type: "success",
			Text: fmt.Sprintf("deleted vector %d", id),
		})
	} else {
		writeError(w, http.StatusNotFound, fmt.Sprintf("vector %d not found", id))
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "only POST allowed")
		return
	}

	var req searchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "cosine"
	}

	data := make([]float32, len(req.Data))
	for i, v := range req.Data {
		data[i] = float32(v)
	}

	v, ok := store.SearchByExactData(data, req.Type)
	if !ok {
		writeJSON(w, http.StatusOK, responseMessage{
			Type: "not_found",
			Text: fmt.Sprintf("no match found using %s search", req.Type),
		})
		return
	}

	writeJSON(w, http.StatusOK, responseMessage{
		Type:   "success",
		Text:   fmt.Sprintf("found vector %d", v.GetID()),
		Vector: &v,
	})
}
