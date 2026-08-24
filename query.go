
func intersect(a, b map[int]
	o

		
func subtract(a, b map[int]bool) map[int]bool {
	out := make(map[int]bool)
	}
	return out
}
func search(query string, idx *Index, snippets map[int]st
	q := strings.TrimSpace(query)
	upper := strings
	var docSet map[int]bool
	isBoolean := false

	switch {
	case strings.Contains(upper, "AND NOT"):
		isBoolean = true
		parts := strings.SplitN(q, "AND NOT", 2)
		if len(parts) == 2 {
			docSet = subtract(docsContaining(strings.TrimSpace(parts[0]), idx),
				docsContaining(strings.TrimSpace(parts[1]), idx))
		}
	case strings.Contains(upper, " AND "):
		isBoolean = true
		parts := strings.Split(q, " AND ")
		docSet = allDocs(idx)
		for _, p := range parts {
			docSet = intersect(docSet, docsContaining(strings.TrimSpace(p), idx))
		}
	case strings.Contains(upper, " OR "):
		isBoolean = true
		parts := strings.Split(q, " OR ")
		docSet = make(map[int]bool)
		for _, p := range parts {
			docSet = union(docSet, docsContaining(strings.TrimSpace(p), idx))
		}
	}

	if isBoolean {
		var results []Result
		for id := range docSet {
			results = append(results, Result{
				DocID:   id,
				Name:    idx.DocNames[id],
				Score:   1.0,
				Snippet: snippets[id],
			})
		}
		sort.Slice(results, func(i, j int) bool { return results[i].Name < results[j].Name })
		return results
	}

	// BM25 ranked
	terms := tokenize(q)
	scores := make(map[int]float64)
	for _, t := range terms {
		for id := range idx.DocLengths {
			scores[id] += bm25(t, id, idx)
		}
	}

	var results []Result
	for id, score := range scores {
		if score > 0 {
			results = append(results, Result{
				DocID:   id,
				Name:    idx.DocNames[id],
				Score:   score,
				Snippet: snippets[id],
			})
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	return results
}
