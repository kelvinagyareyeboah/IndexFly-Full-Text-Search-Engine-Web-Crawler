
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

const (
	docsDir   = "sample_docs"

	port      = ":8080"
)

var (
	idx      *Index
	snippets map[int]string
	tmpl     *template.Template
)

func loadSnippets() {
	snippets = make(map[int]string, idx.N)
	for id, name := range idx.DocNames {
		path := filepath.Join(docsDir, name)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			snippets[id] = scanner.Text()
		}
		f.Close()
	}
}

func reindex() error {
	built, err := buildIndex(docsDir)
	if err != nil {
		return err
	}
	if err := saveIndex(built, indexFile); err != nil {
		return err
	}
	idx = built
	loadSnippets()
	return nil
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		tmpl.Execute(w, map[string]any{"Query": "", "Results": nil, "Total": 0})
		return
	}
	results := search(q, idx, snippets)
	tmpl.Execute(w, map[string]any{
		"Query":   q,
		"Results": results,
		"Total":   len(results),
	})
}

func handleReindex(w http.ResponseWriter, r *http.Request) {
	if err := reindex(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]any{"ok": true, "docs": idx.N, "terms": len(idx.InvertedIndex)})
}

func main() {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
	}
	tmpl = template.Must(
		template.New("index.html").Funcs(funcs).ParseFiles("templates/index.html"),
	)

	var err error
	if _, err = os.Stat(indexFile); os.IsNotExist(err) {
		log.Println("Building index...")
		if err = reindex(); err != nil {
			log.Fatal(err)
		}
	} else {
		idx, err = loadIndex(indexFile)
		if err != nil {
			log.Fatal(err)
		}
		loadSnippets()
	}

	http.HandleFunc("/", handleSearch)
	http.HandleFunc("/reindex", handleReindex)

	fmt.Printf("\n  🔍  Search Engine running → http://localhost%s\n\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
