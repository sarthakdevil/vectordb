package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"vectordb/internal/distances"
	"vectordb/internal/storage"
)

const dataFile = "vectordb.json"

func main() {
	loadDotEnv()
	store := loadStore()

	if len(os.Args) > 1 {
		runCommand(store, os.Args[1], os.Args[2:])
		return
	}

	repl(store)
}

func repl(store *storage.VectorStore) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("> ")
			continue
		}
		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]
		switch cmd {
		case "exit", "quit":
			fmt.Println("bye")
			return
		}
		runCommand(store, cmd, args)
		fmt.Print("> ")
	}
}

func runCommand(store *storage.VectorStore, cmd string, args []string) {
	switch cmd {
	case "add":
		runAdd(store, args)
	case "get":
		runGet(store, args)
	case "list":
		runList(store, args)
	case "delete":
		runDelete(store, args)
	case "search":
		runSearch(store, args)
	case "embed":
		runEmbed(store, args)
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
	}
}

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

func indexType() string {
	if t := os.Getenv("VECTORDB_INDEX"); t == "flat" {
		return "flat"
	}
	return "ivf"
}

func loadStore() *storage.VectorStore {
	store := storage.NewVectorStore(indexType())
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return store
	}
	var vectors []storage.Vector
	if err := json.Unmarshal(data, &vectors); err != nil {
		fmt.Fprintf(os.Stderr, "warning: corrupt data file, starting fresh: %v\n", err)
		return store
	}
	for _, v := range vectors {
		store.Add(v.Data, v.Text)
	}
	return store
}

func saveStore(store *storage.VectorStore) {
	data, err := json.MarshalIndent(store.All(), "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error saving data: %v\n", err)
		return
	}
	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing data file: %v\n", err)
	}
}

func printUsage() {
	fmt.Println(`Usage: vectordb <command> [arguments]

Commands:
  add     Add a new vector
          Usage: vectordb add <text> <f1> <f2> ...

  get     Get a vector by ID
          Usage: vectordb get <id>

  list    List all vectors
          Usage: vectordb list

  delete  Delete a vector by ID
          Usage: vectordb delete <id>

  search  Search for similar vectors (default: cosine)
           Usage: vectordb search [cosine|distance|exact] <text or f1 f2 ...>

  embed   Embed text with Gemini and add to store
           Usage: vectordb embed <text>`)
}

func parseFloats(args []string) ([]float32, error) {
	vals := make([]float32, 0, len(args))
	for _, s := range args {
		v, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid float %q", s)
		}
		vals = append(vals, float32(v))
	}
	return vals, nil
}

func runAdd(store *storage.VectorStore, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: vectordb add <text> <f1> <f2> ...")
		os.Exit(1)
	}

	text := args[0]
	data, err := parseFloats(args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	v := store.Add(data, text)
	saveStore(store)
	fmt.Printf("added vector {id: %d, text: %q, dims: %d}\n", v.GetID(), v.GetText(), len(v.GetData()))
}

func runGet(store *storage.VectorStore, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: vectordb get <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid id: %s\n", args[0])
		os.Exit(1)
	}

	v, ok := store.FindByID(id)
	if !ok {
		fmt.Fprintln(os.Stderr, "not found")
		os.Exit(1)
	}

	printVector(v)
}

func runList(store *storage.VectorStore, args []string) {
	vectors := store.All()
	if len(vectors) == 0 {
		fmt.Println("no vectors in store")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tText\tDimensions")
	for _, v := range vectors {
		fmt.Fprintf(w, "%d\t%s\t%d\n", v.GetID(), v.GetText(), len(v.GetData()))
	}
	w.Flush()
}

func runDelete(store *storage.VectorStore, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: vectordb delete <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid id: %s\n", args[0])
		os.Exit(1)
	}

	if store.DeleteByID(id) {
		saveStore(store)
		fmt.Printf("deleted vector %d\n", id)
	} else {
		fmt.Fprintln(os.Stderr, "not found")
		os.Exit(1)
	}
}

type scoredVector struct {
	v     storage.Vector
	score float32
}

func runSearch(store *storage.VectorStore, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: vectordb search [cosine|distance|exact] <text or f1 f2 ...>")
		return
	}

	searchType := "cosine"
	queryArgs := args

	switch args[0] {
	case "cosine", "distance", "exact":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: vectordb search [cosine|distance|exact] <text or f1 f2 ...>")
			return
		}
		searchType = args[0]
		queryArgs = args[1:]
	}

	data, err := parseFloats(queryArgs)
	if err != nil {
		text := strings.Join(queryArgs, " ")
		fmt.Printf("embedding %q...\n", text)
		vec, err := embedText(text)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return
		}
		data = vec
	}

	all := store.All()
	if len(all) == 0 {
		fmt.Println("no vectors in store")
		return
	}

	var results []scoredVector
	for _, v := range all {
		var score float32
		switch searchType {
		case "cosine":
			score = distances.CosineSimilarity(data, v.GetData())
		case "distance":
			score = -distances.L2Distance(data, v.GetData())
		case "exact":
			if distances.ExactMatch(data, v.GetData()) {
				score = 1
			} else {
				score = 0
			}
		}
		results = append(results, scoredVector{v, score})
	}

	if searchType == "exact" {
		sort.Slice(results, func(i, j int) bool {
			return results[i].v.GetID() < results[j].v.GetID()
		})
	} else {
		sort.Slice(results, func(i, j int) bool {
			return results[i].score > results[j].score
		})
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tScore\tText")
	for _, r := range results {
		if searchType == "exact" && r.score == 0 {
			continue
		}
		fmt.Fprintf(w, "%d\t%.4f\t%s\n", r.v.GetID(), r.score, r.v.GetText())
	}
	w.Flush()
}

func printVector(v storage.Vector) {
	fmt.Printf("id:   %d\n", v.GetID())
	fmt.Printf("text: %s\n", v.GetText())
	fmt.Printf("data: [")
	for i, val := range v.GetData() {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(val)
	}
	fmt.Println("]")
}
