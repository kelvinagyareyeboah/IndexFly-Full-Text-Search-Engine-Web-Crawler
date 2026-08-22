
	f, err := os
	f, err
	}
	defer f.Close()
	var idx Index
	return &idx, json.NewDec
}

// BM25 constants
const k1 = 1.5
const b = 0.75

func bm25(term string, docID int, idx *Index) float64 {
	postings, ok := idx.InvertedIndex[term]
	if !ok {
		return 0
	}
	tf := float64(postings[docID])
	if tf == 0 {
		return 0
	}
	N := float64(idx.N)
	df := float64(len(postings))
	dl := float64(idx.DocLengths[docID])
	idf := math.Log((N-df+0.5)/(df+0.5) + 1)
	tfNorm := (tf * (k1 + 1)) / (tf + k1*(1-b+b*dl/idx.AvgDL))
	return idf * tfNorm
}
