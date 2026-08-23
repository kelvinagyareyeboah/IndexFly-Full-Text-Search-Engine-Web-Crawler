

func main() {
	funcs := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"mul": func(a, b int) int { return a * b },
	}
	tmpl = temp
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
