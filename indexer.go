

func buildIndex(docsDir strin
	entries, err := os.ReadDir(docs
	idx := &Index{
		InvertedIndex: make(map[string
		Doc
		DocN		if e.IsDir() || !strings.HasSuffix(e.Name(), "
		}
		data, err := os.ReadFile(filepath.Join(docsDir, e.Name()))
		if err != nil {
			co		idx.DocNames[docID] = e.Name()
		idx.DocLengths[docID] = len(tokens)

		freq := make(map[string]int)
		for _, t := range tokens {
			freq[t]++
		}
		for term, tf := range freq {
			if idx.InvertedIndex[term] == nil {
				idx.InvertedIndex[term] = make(map[int]int)
			}
			idx.InvertedIndex[term][docID] = tf
		}
		docID++
	}

	idx.N = docID
	total := 0
	for _, l := range idx.DocLengths {
		total += l
	}
	if idx.N > 0 {
		idx.AvgDL = float64(total) / float64(idx.N)
	}
	return idx, nil
}

func saveIndex(idx *Index, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(idx)
}

func loadIndex(path string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var idx Index
	return &idx, json.NewDecoder(f).Decode(&idx)
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
