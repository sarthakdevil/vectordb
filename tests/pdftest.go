package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"vectordb/internal/distances"
	storage "vectordb/internal/storage"
)

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func Pdftest() {
	loadDotEnv()
	ChunkSize := 500
	text, err := loadPDF("test.pdf")
	if err != nil {
		log.Fatal(err)
	}
	chunks := splitTextIntoChunks(text, ChunkSize)
	fmt.Printf("Loaded %d chunks from PDF\n", len(chunks))

	store := storage.NewVectorStore("ivf")

	for i, chunk := range chunks {
		fmt.Printf("Embedding chunk %d/%d...\n", i+1, len(chunks))
		vec, err := embedText(chunk)
		if err != nil {
			log.Printf("Error embedding chunk %d: %v\n", i, err)
			continue
		}
		store.Add(vec, chunk)
	}
	fmt.Println("PDF chunks embedded and stored.")

	questions, err := loadquestions("questions.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nLoaded %d questions.\n", len(questions))

	for _, q := range questions {
		fmt.Printf("\nQuestion: %s\n", q)
		queryVec, err := embedText(q)
		if err != nil {
			log.Printf("Error embedding question: %v\n", err)
			continue
		}
		bestText, bestScore := bestMatch(store, queryVec)
		fmt.Printf("Score: %.4f\n", bestScore)
		fmt.Printf("Best match: %s\n", bestText)
	}
}

func loadPDF(filePath string) (string, error) {
	f, reader, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var text string
	totalPage := reader.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		page := reader.Page(pageIndex)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			log.Println(err)
			continue
		}
		text += content
	}
	return text, nil
}

func splitTextIntoChunks(text string, chunkSize int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var chunks []string
	var builder strings.Builder
	for _, word := range words {
		if builder.Len() > 0 && builder.Len()+len(word)+1 > chunkSize {
			chunks = append(chunks, strings.TrimSpace(builder.String()))
			builder.Reset()
		}
		if builder.Len() > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(word)
	}
	remaining := strings.TrimSpace(builder.String())
	if remaining != "" {
		chunks = append(chunks, remaining)
	}
	return chunks
}

func loadquestions(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var questions []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			questions = append(questions, line)
		}
	}
	return questions, scanner.Err()
}

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

func bestMatch(store *storage.VectorStore, query []float32) (string, float32) {
	all := store.All()
	bestScore := float32(-math.MaxFloat32)
	bestText := ""
	for _, v := range all {
		score := distances.CosineSimilarity(query, v.GetData())
		if score > bestScore {
			bestScore = score
			bestText = v.GetText()
		}
	}
	return bestText, bestScore
}

func embedText(text string) ([]float32, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
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
