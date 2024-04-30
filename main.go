package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

type Document struct {
	ID   int
	Text string
}

type TermFrequency map[string]float64
type InvertedIndex map[string][]int
type TFIDFIndex map[string]map[int]float64

// It takes a string of text as input and tokenizes it into individual words.
// It splits the text into tokens (words) using whitespaces as delimiters and returns a slice of strings containing the tokens.
func tokenize(text string) []string {
	return strings.Fields(text)
}

// It calculates the term frequency (TF) of each token in a given slice of tokens.
// Term frequency is the number of times a term appears in a document divided by the total number of terms in the document.
// It returns a map[string]float64 where the keys are tokens and the values are their corresponding TF scores.
func calculateTermFrequency(tokens []string) TermFrequency {
	tf := make(TermFrequency)
	totalTokens := len(tokens)
	for _, token := range tokens {
		tf[token]++
	}
	for token, freq := range tf {
		tf[token] = freq / float64(totalTokens)
	}
	return tf
}

//	It builds an inverted index from a collection of documents.
//
// An inverted index is a data structure that maps each term (token) to the list of document IDs in which it appears.
// It returns a map[string][]int where the keys are terms and the values are slices of document IDs.
func buildInvertedIndex(docs []Document) InvertedIndex {
	index := make(InvertedIndex)
	for _, doc := range docs {
		tokens := tokenize(doc.Text)
		for _, token := range tokens {
			index[token] = append(index[token], doc.ID)
		}
	}
	return index
}

// It calculates the inverse document frequency (IDF) of each term in an inverted index.
// IDF is a measure of how important a term is across all documents in a corpus. It is calculated as the logarithm of the total number of documents divided by the number of documents containing the term.
// It returns a map[string]float64 where the keys are terms and the values are their corresponding IDF scores.
func calculateIDF(index InvertedIndex, totalDocs int) map[string]float64 {
	idf := make(map[string]float64)
	for term, docIDs := range index {
		idf[term] = math.Log(float64(totalDocs) / float64(len(docIDs)))
	}
	return idf
}

// It builds a TF-IDF index from a collection of documents and their corresponding inverted index.
// TF-IDF (Term Frequency-Inverse Document Frequency) is a numerical statistic that reflects the importance of a term in a document relative to a corpus.
// It returns a map[string]map[int]float64 where the outer map's keys are terms and the inner map's keys are document IDs, and the values are their corresponding TF-IDF scores.
func buildTFIDFIndex(docs []Document, invertedIndex InvertedIndex) TFIDFIndex {
	tfidfIndex := make(TFIDFIndex)
	totalDocs := len(docs)

	idf := calculateIDF(invertedIndex, totalDocs)

	for _, doc := range docs {
		tokens := tokenize(doc.Text)
		docTF := calculateTermFrequency(tokens)
		for token, tf := range docTF {
			idfValue := idf[token]
			if _, ok := tfidfIndex[token]; !ok {
				tfidfIndex[token] = make(map[int]float64)
			}
			tfidfIndex[token][doc.ID] = tf * idfValue
		}
	}

	return tfidfIndex
}

// It performs a TF-IDF based search for a given query string.
// It calculates the TF-IDF scores for each term in the query and aggregates them across documents to produce search results.
// It returns a map[int]float64 where the keys are document IDs and the values are their corresponding TF-IDF scores.
func searchTFIDF(tfidfIndex TFIDFIndex, query string) map[int]float64 {
	terms := tokenize(query)
	queryTF := calculateTermFrequency(terms)

	result := make(map[int]float64)

	for term, tf := range queryTF {
		if docIDs, ok := tfidfIndex[term]; ok {
			for docID, tfidf := range docIDs {
				result[docID] += tf * tfidf // accumulate TF-IDF score
			}
		}
	}

	return result
}

// It ranks search results based on their TF-IDF scores.
// It sorts the search results by TF-IDF score in descending order.
// It returns a slice of document IDs sorted by their TF-IDF scores.
func rankSearchResults(results map[int]float64) []int {
	type result struct {
		docID int
		score float64
	}

	var rankedResults []result
	for docID, score := range results {
		rankedResults = append(rankedResults, result{docID, score})
	}

	sort.Slice(rankedResults, func(i, j int) bool {
		return rankedResults[i].score > rankedResults[j].score
	})

	var rankedDocIDs []int
	for _, res := range rankedResults {
		rankedDocIDs = append(rankedDocIDs, res.docID)
	}

	return rankedDocIDs
}

func main() {
	// Example documents
	docs := []Document{
		{ID: 1, Text: "This is a document about Go programming."},
		{ID: 2, Text: "Go is a statically typed, compiled programming language."},
		{ID: 3, Text: "Elasticsearch is a distributed, RESTful search and analytics engine. Go"},
	}

	// Build inverted index and TF-IDF index
	invertedIndex := buildInvertedIndex(docs)
	tfidfIndex := buildTFIDFIndex(docs, invertedIndex)

	// Perform search
	query := "Go"
	searchResults := searchTFIDF(tfidfIndex, query)

	// Rank search results
	rankedResults := rankSearchResults(searchResults)

	// Print results
	fmt.Printf("Search results for query '%s':\n", query)
	for _, docID := range rankedResults {
		for _, doc := range docs {
			if doc.ID == docID {
				fmt.Printf("%s\n", doc.Text)
				break
			}
		}
	}
}
