i
		for _, p := range parts {
			docSet = union(docSet,
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
