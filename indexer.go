
	"path/
	"regexp"
	"strings"
)

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "it": true,
	"in": true, "on": true, "at": true, "to": true, "of": true,
	"and": true, "or": true, "for"": true, "that": true,
	"this": true, "are": true, "was": true, "be": true, "by": true, "from": tru
var tokenRe = regexp.MustCompile
type Index struct {
	InvertedIndex map[string]map[int]int `json:"inverted_index"` // term -> docID -> tf
	DocLengths    map[int]int            `json:"doc_lengths"`
	DocNames      map[int]string         `json:"doc_names"`
	N             int                    `json:"n"`
	AvgDL         float64                `json:"avg_dl"`
}

func tokenize(text string) []string {
	words := tokenRe.FindAllString(strings.ToLower(text), -1)
	out := words[:0]
	for _, w := range words {
		if !stopWords[w] {
			out = append(out, w)
		}
	}
	return out
}

func buildIndex(docsDir string) (*Index, error) {
	entries, err := os.ReadDir(docsDir)
	if err != nil {
		return nil, err
	}

	idx := &Index{
		InvertedIndex: make(map[string]map[int]int),
		DocLengths:    make(map[int]int),
		DocNames:      make(map[int]string),
	}

	docID := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(docsDir, e.Name()))
		if err != nil {
			continue
		}
		tokens := tokenize(string(data))
		idx.DocNames[docID] = e.Name()
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
