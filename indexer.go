package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

//
// ------------------------------------------------------------
// CONFIGURATION
// ------------------------------------------------------------
//

const (
	k1 = 1.5
	b  = 0.75

	titleBoost  = 1.25
	phraseBoost = 1.50
	fuzzyBoost  = 0.35

	defaultTopK = 10
)

var stopWords = map[string]bool{
	"a": true, "an": true, "the": true,
	"is": true, "it": true, "in": true,
	"on": true, "at": true, "to": true,
	"of": true, "and": true, "or": true,
	"for": true, "with": true, "that": true,
	"this": true, "are": true, "was": true,
	"be": true, "by": true, "from": true,
	"as": true, "but": true, "not": true,
	"into": true, "about": true, "over": true,
	"after": true, "before": true, "between": true,
	"through": true, "during": true, "without": true,
	"within": true, "can": true, "could": true,
	"would": true, "should": true, "will": true,
	"have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true,
}

var tokenRe = regexp.MustCompile(`[a-zA-Z0-9]+`)

//
// ------------------------------------------------------------
// DATA STRUCTURES
// ------------------------------------------------------------
//

type Index struct {
	InvertedIndex map[string]map[int]int `json:"inverted_index"`
	DocLengths    map[int]int            `json:"doc_lengths"`
	DocNames      map[int]string         `json:"doc_names"`
	Documents    map[int]string         `json:"documents"`

	N     int     `json:"n"`
	AvgDL float64 `json:"avg_dl"`
}

type SearchResult struct {
	DocID      int
	Document   string
	Score      float64
	Snippet    string
	Matched    []string
}

type SearchEngine struct {
	Index *Index
	Text  map[int]string
}

//
// ------------------------------------------------------------
// TOKENIZATION
// ------------------------------------------------------------
//

func tokenize(text string) []string {
	text = strings.ToLower(text)

	words := tokenRe.FindAllString(text, -1)

	out := make([]string, 0, len(words))

	for _, word := range words {
		if stopWords[word] {
			continue
		}

		word = stem(word)

		if word == "" {
			continue
		}

		out = append(out, word)
	}

	return out
}

//
// ------------------------------------------------------------
// SIMPLE STEMMER
// ------------------------------------------------------------
//

func stem(word string) string {

	if len(word) <= 3 {
		return word
	}

	// plural
	if strings.HasSuffix(word, "ies") && len(word) > 4 {
		return word[:len(word)-3] + "y"
	}

	if strings.HasSuffix(word, "sses") {
		return word[:len(word)-2]
	}

	if strings.HasSuffix(word, "s") &&
		!strings.HasSuffix(word, "ss") {
		word = word[:len(word)-1]
	}

	// ing
	if strings.HasSuffix(word, "ing") && len(word) > 5 {
		word = word[:len(word)-3]

		if len(word) > 2 &&
			word[len(word)-1] == word[len(word)-2] {
			word = word[:len(word)-1]
		}
	}

	// ed
	if strings.HasSuffix(word, "ed") && len(word) > 4 {
		word = word[:len(word)-2]
	}

	// ly
	if strings.HasSuffix(word, "ly") && len(word) > 4 {
		word = word[:len(word)-2]
	}

	// ment
	if strings.HasSuffix(word, "ment") && len(word) > 6 {
		word = word[:len(word)-4]
	}

	return word
}

//
// ------------------------------------------------------------
// INDEX BUILDING
// ------------------------------------------------------------
//

func buildIndex(docsDir string) (*SearchEngine, error) {

	entries, err := os.ReadDir(docsDir)
	if err != nil {
		return nil, err
	}

	idx := &Index{
		InvertedIndex: make(map[string]map[int]int),
		DocLengths:    make(map[int]int),
		DocNames:      make(map[int]string),
		Documents:     make(map[int]string),
	}

	engine := &SearchEngine{
		Index: idx,
		Text:  make(map[int]string),
	}

	type document struct {
		id   int
		name string
		text string
	}

	var documents []document

	docID := 0

	for _, entry := range entries {

		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(
			strings.ToLower(entry.Name()),
			".txt",
		) {
			continue
		}

		path := filepath.Join(
			docsDir,
			entry.Name(),
		)

		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		documents = append(documents, document{
			id:   docID,
			name: entry.Name(),
			text: string(data),
		})

		docID++
	}

	//
	// Process documents concurrently.
	//

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, doc := range documents {

		wg.Add(1)

		go func(d document) {

			defer wg.Done()

			tokens := tokenize(d.text)

			freq := make(map[string]int)

			for _, token := range tokens {
				freq[token]++
			}

			mu.Lock()

			idx.DocNames[d.id] = d.name
			idx.Documents[d.id] = d.text
			idx.DocLengths[d.id] = len(tokens)

			engine.Text[d.id] = d.text

			for term, tf := range freq {

				if idx.InvertedIndex[term] == nil {
					idx.InvertedIndex[term] =
						make(map[int]int)
				}

				idx.InvertedIndex[term][d.id] = tf
			}

			mu.Unlock()

		}(doc)
	}

	wg.Wait()

	idx.N = len(documents)

	totalLength := 0

	for _, length := range idx.DocLengths {
		totalLength += length
	}

	if idx.N > 0 {
		idx.AvgDL =
			float64(totalLength) /
				float64(idx.N)
	}

	return engine, nil
}

//
// ------------------------------------------------------------
// SAVE INDEX
// ------------------------------------------------------------
//

func saveIndex(
	engine *SearchEngine,
	path string,
) error {

	data, err := json.MarshalIndent(
		engine.Index,
		"",
		"  ",
	)

	if err != nil {
		return err
	}

	return os.WriteFile(
		path,
		data,
		0644,
	)
}

//
// ------------------------------------------------------------
// LOAD INDEX
// ------------------------------------------------------------
//

func loadIndex(
	path string,
) (*SearchEngine, error) {

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var idx Index

	err = json.Unmarshal(
		data,
		&idx,
	)

	if err != nil {
		return nil, err
	}

	engine := &SearchEngine{
		Index: &idx,
		Text:  make(map[int]string),
	}

	for id, text := range idx.Documents {
		engine.Text[id] = text
	}

	return engine, nil
}

//
// ------------------------------------------------------------
// BM25
// ------------------------------------------------------------
//

func bm25(
	term string,
	docID int,
	idx *Index,
) float64 {

	postings, exists :=
		idx.InvertedIndex[term]

	if !exists {
		return 0
	}

	tfValue, exists :=
		postings[docID]

	if !exists || tfValue == 0 {
		return 0
	}

	tf := float64(tfValue)

	N := float64(idx.N)

	df := float64(len(postings))

	dl := float64(
		idx.DocLengths[docID],
	)

	if idx.AvgDL == 0 {
		return 0
	}

	idf := math.Log(
		((N-df+0.5)/
			(df+0.5)) + 1,
	)

	tfNorm :=
		(tf * (k1 + 1)) /
			(tf +
				k1*
					(1-b+
						b*(dl/idx.AvgDL)))

	return idf * tfNorm
}

//
// ------------------------------------------------------------
// QUERY PARSING
// ------------------------------------------------------------
//

type Query struct {
	Terms   []string
	Phrases [][]string
}

func parseQuery(query string) Query {

	result := Query{}

	//
	// Extract quoted phrases.
	//

	phraseRe :=
		regexp.MustCompile(`"([^"]+)"`)

	phrases :=
		phraseRe.FindAllStringSubmatch(
			query,
			-1,
		)

	for _, phrase := range phrases {

		tokens :=
			tokenize(phrase[1])

		if len(tokens) > 0 {
			result.Phrases =
				append(
					result.Phrases,
					tokens,
				)
		}
	}

	//
	// Remove phrases from normal query.
	//

	query =
		phraseRe.ReplaceAllString(
			query,
			"",
		)

	result.Terms =
		tokenize(query)

	return result
}

//
// ------------------------------------------------------------
// PHRASE MATCHING
// ------------------------------------------------------------
//

func containsPhrase(
	text string,
	phrase []string,
) bool {

	if len(phrase) == 0 {
		return false
	}

	tokens :=
		tokenize(text)

	if len(tokens) < len(phrase) {
		return false
	}

	for i := 0; i <=
		len(tokens)-len(phrase); i++ {

		match := true

		for j := range phrase {

			if tokens[i+j] !=
				phrase[j] {

				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

//
// ------------------------------------------------------------
// FUZZY MATCHING
// ------------------------------------------------------------
//

func levenshtein(a, b string) int {

	if a == b {
		return 0
	}

	if len(a) == 0 {
		return len(b)
	}

	if len(b) == 0 {
		return len(a)
	}

	prev :=
		make([]int, len(b)+1)

	curr :=
		make([]int, len(b)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {

		curr[0] = i

		for j := 1; j <= len(b); j++ {

			cost := 0

			if a[i-1] != b[j-1] {
				cost = 1
			}

			curr[j] = min(
				curr[j-1]+1,
				prev[j]+1,
				prev[j-1]+cost,
			)
		}

		prev, curr =
			curr, prev
	}

	return prev[len(b)]
}

func min(a, b, c int) int {

	if a < b && a < c {
		return a
	}

	if b < c {
		return b
	}

	return c
}

func fuzzyTerms(
	term string,
	idx *Index,
) []string {

	results := []string{}

	for indexedTerm :=
		range idx.InvertedIndex {

		distance :=
			levenshtein(
				term,
				indexedTerm,
			)

		threshold := 1

		if len(term) >= 6 {
			threshold = 2
		}

		if distance <= threshold {
			results =
				append(
					results,
					indexedTerm,
				)
		}
	}

	return results
}

//
// ------------------------------------------------------------
// SEARCH
// ------------------------------------------------------------
//

func (e *SearchEngine) Search(
	queryText string,
	topK int,
) []SearchResult {

	if topK <= 0 {
		topK = defaultTopK
	}

	query :=
		parseQuery(queryText)

	scores :=
		make(map[int]float64)

	matchedTerms :=
		make(map[int]map[string]bool)

	//
	// Exact query terms.
	//

	for _, term := range query.Terms {

		postings :=
			e.Index.InvertedIndex[term]

		for docID := range postings {

			score :=
				bm25(
					term,
					docID,
					e.Index,
				)

			scores[docID] += score

			if matchedTerms[docID] == nil {
				matchedTerms[docID] =
					make(map[string]bool)
			}

			matchedTerms[docID][term] = true
		}

		//
		// Fuzzy matching.
		//

		if len(postings) == 0 {

			alternatives :=
				fuzzyTerms(
					term,
					e.Index,
				)

			for _, alternative :=
				range alternatives {

				for docID :=
					range e.Index.InvertedIndex[alternative] {

					score :=
						bm25(
							alternative,
							docID,
							e.Index,
						)

					scores[docID] +=
						score * fuzzyBoost

					if matchedTerms[docID] == nil {
						matchedTerms[docID] =
							make(map[string]bool)
					}

					matchedTerms[docID][alternative] =
						true
				}
			}
		}
	}

	//
	// Phrase boosting.
	//

	for docID := range scores {

		text :=
			e.Text[docID]

		for _, phrase :=
			range query.Phrases {

			if containsPhrase(
				text,
				phrase,
			) {

				scores[docID] *=
					phraseBoost
			}
		}
	}

	//
	// Document-name/title boosting.
	//

	for docID := range scores {

		name :=
			strings.ToLower(
				e.Index.DocNames[docID],
			)

		for _, term :=
			range query.Terms {

			if strings.Contains(
				name,
				term,
			) {
				scores[docID] *=
					titleBoost
			}
		}
	}

	//
	// Convert to result objects.
	//

	results :=
		make([]SearchResult, 0, len(scores))

	for docID, score :=
		range scores {

		if score <= 0 {
			continue
		}

		matched := []string{}

		for term :=
			range matchedTerms[docID] {

			matched =
				append(
					matched,
					term,
				)
		}

		sort.Strings(matched)

		results =
			append(
				results,
				SearchResult{
					DocID:    docID,
					Document: e.Index.DocNames[docID],
					Score:    score,
					Snippet:  makeSnippet(
						e.Text[docID],
						matched,
					),
					Matched: matched,
				},
			)
	}

	//
	// Sort highest score first.
	//

	sort.Slice(
		results,
		func(i, j int) bool {
			return results[i].Score >
				results[j].Score
		},
	)

	if len(results) > topK {
		results =
			results[:topK]
	}

	return results
}

//
// ------------------------------------------------------------
// SNIPPET GENERATION
// ------------------------------------------------------------
//

func makeSnippet(
	text string,
	terms []string,
) string {

	if text == "" {
		return ""
	}

	words :=
		strings.Fields(text)

	if len(words) <= 30 {
		return strings.Join(words, " ")
	}

	position := 0

	for i, word :=
		range words {

		clean :=
			strings.ToLower(
				strings.TrimFunc(
					word,
					func(r rune) bool {
						return !unicode.IsLetter(r) &&
							!unicode.IsNumber(r)
					},
				),
			)

		for _, term :=
			range terms {

			if strings.Contains(
				clean,
				term,
			) {
				position = i
				break
			}
		}
	}

	start :=
		position - 10

	if start < 0 {
		start = 0
	}

	end :=
		start + 30

	if end > len(words) {
		end = len(words)
	}

	snippet :=
		strings.Join(
			words[start:end],
			" ",
		)

	return highlight(
		snippet,
		terms,
	)
}

//
// ------------------------------------------------------------
// HIGHLIGHTING
// ------------------------------------------------------------
//

func highlight(
	text string,
	terms []string,
) string {

	for _, term := range terms {

		re :=
			regexp.MustCompile(
				`(?i)\b` +
					regexp.QuoteMeta(term) +
					`\b`,
			)

		text =
			re.ReplaceAllString(
				text,
				"[$0]",
			)
	}

	return text
}

//
// ------------------------------------------------------------
// INDEX STATISTICS
// ------------------------------------------------------------
//

func (e *SearchEngine) Stats() {

	fmt.Println()
	fmt.Println("========== SEARCH INDEX ==========")
	fmt.Println(
		"Documents:",
		e.Index.N,
	)

	fmt.Println(
		"Unique terms:",
		len(e.Index.InvertedIndex),
	)

	fmt.Printf(
		"Average document length: %.2f\n",
		e.Index.AvgDL,
	)

	fmt.Println(
		"Total indexed terms:",
		totalTerms(e.Index),
	)

	fmt.Println(
		"==================================",
	)
	fmt.Println()
}

func totalTerms(idx *Index) int {

	total := 0

	for _, length :=
		range idx.DocLengths {

		total += length
	}

	return total
}

//
// ------------------------------------------------------------
// SEARCH DISPLAY
// ------------------------------------------------------------
//

func printResults(
	results []SearchResult,
) {

	if len(results) == 0 {

		fmt.Println(
			"No matching documents found.",
		)

		return
	}

	fmt.Println()

	for i, result :=
		range results {

		fmt.Printf(
			"%d. %s\n",
			i+1,
			result.Document,
		)

		fmt.Printf(
			"   Score: %.4f\n",
			result.Score,
		)

		fmt.Printf(
			"   Matched: %s\n",
			strings.Join(
				result.Matched,
				", ",
			),
		)

		fmt.Printf(
			"   %s\n",
			result.Snippet,
		)

		fmt.Println()
	}
}

//
// ------------------------------------------------------------
// COMMAND LINE INTERFACE
// ------------------------------------------------------------
//

func usage() {

	fmt.Println(`
BM25 SEARCH ENGINE

Usage:

  search index <documents-folder> <index-file>
  search query <index-file> <query>
  search stats <index-file>

Examples:

  search index ./documents index.json

  search query index.json "machine learning"

  search query index.json artificial intelligence

  search stats index.json
`)
}

//
// ------------------------------------------------------------
// MAIN
// ------------------------------------------------------------
//

func main() {

	if len(os.Args) < 2 {
		usage()
		return
	}

	command :=
		strings.ToLower(
			os.Args[1],
		)

	switch command {

	case "index":

		if len(os.Args) < 4 {

			fmt.Println(
				"Usage: search index <documents-folder> <index-file>",
			)

			return
		}

		docsDir := os.Args[2]
		indexFile := os.Args[3]

		fmt.Println(
			"Building search index...",
		)

		engine, err :=
			buildIndex(docsDir)

		if err != nil {

			fmt.Println(
				"Indexing error:",
				err,
			)

			return
		}

		err =
			saveIndex(
				engine,
				indexFile,
			)

		if err != nil {

			fmt.Println(
				"Could not save index:",
				err,
			)

			return
		}

		fmt.Println(
			"Index successfully created.",
		)

		engine.Stats()

	case "query":

		if len(os.Args) < 4 {

			fmt.Println(
				"Usage: search query <index-file> <query>",
			)

			return
		}

		indexFile := os.Args[2]

		query :=
			strings.Join(
				os.Args[3:],
				" ",
			)

		engine, err :=
			loadIndex(indexFile)

		if err != nil {

			fmt.Println(
				"Could not load index:",
				err,
			)

			return
		}

		results :=
			engine.Search(
				query,
				defaultTopK,
			)

		fmt.Println(
			"Query:",
			query,
		)

		printResults(results)

	case "stats":

		if len(os.Args) < 3 {

			fmt.Println(
				"Usage: search stats <index-file>",
			)

			return
		}

		engine, err :=
			loadIndex(os.Args[2])

		if err != nil {

			fmt.Println(
				"Could not load index:",
				err,
			)

			return
		}

		engine.Stats()

	default:

		usage()
	}
}

//
// ------------------------------------------------------------
// END
// ------------------------------------------------------------
