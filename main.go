func(a, b int) int { return a + b },
		"mul": func(a, b int) int { ret
	tmpl = temp
	
	if t(indexFile); os.IsNotExist(err) {
		log.Println("Bui
		if err = reindex();
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
