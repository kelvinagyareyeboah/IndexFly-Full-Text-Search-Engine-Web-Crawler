package main

import (
	"sort"
	"strings"
)

// Result represents one search result returned to the user.
type Result struct {
	DocID   int
	Name    string
	Score   float64
	Snippet string
}

// allDocs returns a set containing every document ID
// that currently exists in the search index.
func allDocs(idx *Index) map[int]bool {

	set := make(
		map[int]bool,
		idx.N,
	)

	for id := range idx.DocLengths {

		set[id] = true
	}

	return set
}

// docsContaining returns all documents that contain
// the specified search term.
func docsContaining(
	term string,
	idx *Index,
) map[int]bool {

	set := make(map[int]bool)

	term = strings.ToLower(term)

	for id := range idx.InvertedIndex[term] {

		set[id] = true
	}

	return set
}

// intersect returns only the document IDs
// that exist in both sets.
func intersect(
	a map[int]bool,
	b map[int]bool,
) map[int]bool {

	out := make(map[int]bool)

	for id := range a {

		if b[id] {

			out[id] = true
		}
	}

	return out
}

// union combines two document sets.
//
// Any document that exists in either set
// will be included in the result.
func union(
	a map[int]bool,
	b map[int]bool,
) map[int]bool {

	out := make(map[int]bool)

	for id := range a {

		out[id] = true
	}

	for id := range b {

		out[id] = true
	}

	return out
}

// subtract removes every document in b
// from the documents contained in a.
func subtract(
	a map[int]bool,
	b map[int]bool,
) map[int]bool {

	out := make(map[int]bool)

	for id := range a {

		if !b[id] {

			out[id] = true
		}
	}

	return out
}

// search processes a user's query.
//
// It supports:
//
//   - Normal keyword searches
//   - BM25 ranking
//   - AND queries
//   - OR queries
//   - AND NOT queries
//
// Normal searches are ranked using BM25,
// while Boolean searches return matching
// documents without BM25 ranking.
func search(
	query string,
	idx *Index,
	snippets map[int]string,
) []Result {

	// Remove unnecessary whitespace
	// from the user's query.
	q := strings.TrimSpace(query)

	// Create an uppercase version so that
	// Boolean operators can be detected
	// regardless of their capitalization.
	upper := strings.ToUpper(q)

	var docSet map[int]bool

	isBoolean := false

	// --------------------------------------------------
	// AND NOT
	// --------------------------------------------------
	//
	// Example:
	//
	//   machine AND NOT learning
	//
	// This returns documents containing
	// "machine" but not "learning".
	//
	switch {

	case strings.Contains(
		upper,
		"AND NOT",
	):

		isBoolean = true

		parts := strings.SplitN(
			q,
			"AND NOT",
			2,
		)

		if len(parts) == 2 {

			leftTerm := strings.TrimSpace(
				parts[0],
			)

			rightTerm := strings.TrimSpace(
				parts[1],
			)

			leftDocs := docsContaining(
				leftTerm,
				idx,
			)

			rightDocs := docsContaining(
				rightTerm,
				idx,
			)

			docSet = subtract(
				leftDocs,
				rightDocs,
			)
		}

	// --------------------------------------------------
	// AND
	// --------------------------------------------------
	//
	// Example:
	//
	//   machine AND learning
	//
	// A document must contain every
	// search term.
	//
	case strings.Contains(
		upper,
		" AND ",
	):

		isBoolean = true

		parts := strings.Split(
			q,
			" AND ",
		)

		// Start with every document.
		docSet = allDocs(idx)

		for _, part := range parts {

			term := strings.TrimSpace(
				part,
			)

			matchingDocs := docsContaining(
				term,
				idx,
			)

			docSet = intersect(
				docSet,
				matchingDocs,
			)
		}

	// --------------------------------------------------
	// OR
	// --------------------------------------------------
	//
	// Example:
	//
	//   machine OR learning
	//
	// A document only needs to contain
	// at least one of the terms.
	//
	case strings.Contains(
		upper,
		" OR ",
	):

		isBoolean = true

		parts := strings.Split(
			q,
			" OR ",
		)

		docSet = make(
			map[int]bool,
		)

		for _, part := range parts {

			term := strings.TrimSpace(
				part,
			)

			matchingDocs := docsContaining(
				term,
				idx,
			)

			docSet = union(
				docSet,
				matchingDocs,
			)
		}
	}

	// --------------------------------------------------
	// BOOLEAN SEARCH RESULTS
	// --------------------------------------------------
	//
	// Boolean searches are not ranked using BM25.
	// Every matching document receives a score of 1.0.
	//
	if isBoolean {

		var results []Result

		for id := range docSet {

			result := Result{
				DocID: id,

				Name: idx.DocNames[id],

				Score: 1.0,

				Snippet: snippets[id],
			}

			results = append(
				results,
				result,
			)
		}

		// Sort Boolean results alphabetically
		// by document name.
		sort.Slice(
			results,
			func(i, j int) bool {

				return results[i].Name <
					results[j].Name
			},
		)

		return results
	}

	// --------------------------------------------------
	// BM25 RANKED SEARCH
	// --------------------------------------------------
	//
	// If the query is not Boolean,
	// tokenize it into individual terms
	// and calculate a BM25 score for each term.
	//
	terms := tokenize(q)

	scores := make(
		map[int]float64,
	)

	// Calculate the score for every
	// search term across the documents.
	for _, term := range terms {

		for id := range idx.DocLengths {

			score := bm25(
				term,
				id,
				idx,
			)

			scores[id] += score
		}
	}

	// --------------------------------------------------
	// CREATE RESULTS
	// --------------------------------------------------

	var results []Result

	for id, score := range scores {

		// Ignore documents that received
		// no BM25 score.
		if score <= 0 {
			continue
		}

		result := Result{
			DocID: id,

			Name: idx.DocNames[id],

			Score: score,

			Snippet: snippets[id],
		}

		results = append(
			results,
			result,
		)
	}

	// --------------------------------------------------
	// SORT RESULTS BY BM25 SCORE
	// --------------------------------------------------
	//
	// Documents with higher scores
	// appear first.
	//
	sort.Slice(
		results,
		func(i, j int) bool {

			return results[i].Score >
				results[j].Score
		},
	)

	return results
}
