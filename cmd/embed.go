package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"vectordb/internal/storage"
)

type geminiRequest struct {
	Content struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"content"`
}

type geminiError struct {
	Message string `json:"message"`
}

type geminiResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
	Error *geminiError `json:"error,omitempty"`
}

type geminiErrorResponse struct {
	Error *geminiError `json:"error"`
}

func embedText(text string) ([]float32, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set (get one free at https://aistudio.google.com/apikey)")
	}

	model := os.Getenv("GEMINI_EMBED_MODEL")
	if model == "" {
		model = "gemini-embedding-001"
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent", model)

	body := geminiRequest{}
	body.Content.Parts = []struct {
		Text string `json:"text"`
	}{{Text: text}}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Gemini API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		var errResp geminiErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != nil {
			return nil, fmt.Errorf("Gemini API error: %s", errResp.Error.Message)
		}
		return nil, fmt.Errorf("Gemini API returned status %d", resp.StatusCode)
	}

	var result geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	vec := make([]float32, len(result.Embedding.Values))
	for i, v := range result.Embedding.Values {
		vec[i] = float32(v)
	}
	return vec, nil
}

func runEmbed(store *storage.VectorStore, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: vectordb embed <text>")
		fmt.Fprintln(os.Stderr, "       vectordb embed \"<text with spaces>\"")
		return
	}

	text := strings.Join(args, " ")
	fmt.Printf("embedding %q...\n", text)

	vec, err := embedText(text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return
	}

	v := store.Add(vec, text)
	saveStore(store)
	fmt.Printf("added vector {id: %d, text: %q, dims: %d}\n", v.GetID(), v.GetText(), len(v.GetData()))
}
