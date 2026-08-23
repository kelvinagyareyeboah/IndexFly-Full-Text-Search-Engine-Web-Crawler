nt]string, id
		path := filep
func handleSearch(w http.Resp
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		tmpl.Execute(w, map[string]any{"Query": "", "Results": nil, "Total": 0})
		return
	}
	results := search(q, 
	tmpl.Execute(w, map[string]any{
	
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
