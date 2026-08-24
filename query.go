
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
