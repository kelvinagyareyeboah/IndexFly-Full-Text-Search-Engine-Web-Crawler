package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	docsDir   = "sample_docs"
	indexFile = "index.json"
	port      = ":8080"
)

var (
	idx      *Index
	snippets map[int]string
	tmpl     *template.Template
)

// loadSnippets reads the first line of every indexed document
// and stores it as a short preview for the search results.
func loadSnippets() {

	snippets = make(map[int]string, idx.N)

	for id, name := range idx.DocNames {

		path := filepath.Join(docsDir, name)

		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)

		if scanner.Scan() {
			snippets[id] = scanner.Text()
		}

		file.Close()
	}
}

// reindex rebuilds the search index from the documents directory.
// The newly created index is then saved to disk.
func reindex() error {

	builtIndex, err := buildIndex(docsDir)

	if err != nil {
		return err
	}

	err = saveIndex(builtIndex, indexFile)

	if err != nil {
		return err
	}

	idx = builtIndex

	loadSnippets()

	return nil
}

// handleSearch handles search requests from the web interface.
func handleSearch(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query().Get("q")

	query = strings.TrimSpace(query)

	// If the user has not entered a search query,
	// display the normal search page.
	if query == "" {

		data := map[string]any{
			"Query":   "",
			"Results": nil,
			"Total":   0,
		}

		err := tmpl.Execute(w, data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		return
	}

	// Perform the actual search.
	results := search(
		query,
		idx,
		snippets,
	)

	// Prepare the data that will be sent to the HTML template.
	data := map[string]any{
		"Query":   query,
		"Results": results,
		"Total":   len(results),
	}

	// Render the search results page.
	err := tmpl.Execute(w, data)

	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}
}

// handleReindex allows the search index to be rebuilt
// through an HTTP request.
func handleReindex(w http.ResponseWriter, r *http.Request) {

	err := reindex()

	if err != nil {

		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)

		return
	}

	// Return information about the newly created index.
	response := map[string]any{
		"ok":    true,
		"docs":  idx.N,
		"terms": len(idx.InvertedIndex),
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(response)
}

func main() {

	// Template helper functions.
	funcs := template.FuncMap{

		"add": func(a, b int) int {
			return a + b
		},

		"mul": func(a, b int) int {
			return a * b
		},
	}

	// Load the HTML template.
	tmpl = template.Must(
		template.
			New("index.html").
			Funcs(funcs).
			ParseFiles("templates/index.html"),
	)

	var err error

	// Check whether the search index already exists.
	_, err = os.Stat(indexFile)

	if os.IsNotExist(err) {

		// No index exists, so create one.
		log.Println("Building index...")

		err = reindex()

		if err != nil {
			log.Fatal(err)
		}

	} else {

		// An index already exists.
		// Load it instead of rebuilding it.
		idx, err = loadIndex(indexFile)

		if err != nil {
			log.Fatal(err)
		}

		// Load document snippets into memory.
		loadSnippets()
	}

	// Register the main search page.
	http.HandleFunc(
		"/",
		handleSearch,
	)

	// Register the index rebuilding endpoint.
	http.HandleFunc(
		"/reindex",
		handleReindex,
	)

	// Display server information.
	fmt.Printf(
		"\n  🔍  Search Engine running → http://localhost%s\n\n",
		port,
	)

	// Start the HTTP server.
	log.Fatal(
		http.ListenAndServe(
			port,
			nil,
		),
	)
}
	
