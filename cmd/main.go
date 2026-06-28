package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"vectordb/internal/storage"
)

const dataFile = "vectordb.json"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	store := loadStore()

	switch os.Args[1] {
	case "add":
		runAdd(store, os.Args[2:])
	case "get":
		runGet(store, os.Args[2:])
	case "list":
		runList(store, os.Args[2:])
	case "delete":
		runDelete(store, os.Args[2:])
	case "search":
		runSearch(store, os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
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

  search  Search for similar vectors
          Usage: vectordb search <cosine|distance|exact> <f1> <f2> ...`)
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

func runSearch(store *storage.VectorStore, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: vectordb search <cosine|distance|exact> <f1> <f2> ...")
		os.Exit(1)
	}

	searchType := args[0]
	data, err := parseFloats(args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	v, ok := store.SearchByExactData(data, searchType)
	if !ok {
		fmt.Println("no match found")
		return
	}

	printVector(v)
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
