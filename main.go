package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
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

var docs = []Document{
	{ID: 1, Text: "This is a document about Go programming."},
	{ID: 2, Text: "Go is a statically typed, compiled programming language."},
	{ID: 3, Text: "Elasticsearch is a distributed, RESTful search and analytics engine. Go"},
	{ID: 4, Text: "A programming language is a computer language intended to formulate algorithms and produce computer programs that apply them."},
	{ID: 5, Text: "Pair programming is an agile working method which is based on collaboration between two developers on the same workstation for the creation and coding of a computer program."},
	{ID: 6, Text: "An API is a set of definitions and protocols that facilitates the creation and integration of application software."},
	{ID: 7, Text: "Efficient Testing Strategies for Go Functions Handling Large Data Inserts into PostgreSQL Tables"},
}

var invertedIndex InvertedIndex
var tfidfIndex TFIDFIndex

// Initialize indexes
func init() {
	invertedIndex = buildInvertedIndex(docs)
	tfidfIndex = buildTFIDFIndex(docs, invertedIndex)
}

// Tokenize text into words and normalize them to lowercase
func tokenize(text string) []string {
	words := strings.Fields(text)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return words
}

// Calculate term frequency
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

// Build inverted index
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

// Calculate IDF
func calculateIDF(index InvertedIndex, totalDocs int) map[string]float64 {
	idf := make(map[string]float64)
	for term, docIDs := range index {
		idf[term] = math.Log(float64(totalDocs) / float64(len(docIDs)))
	}
	return idf
}

// Build TF-IDF index
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

// Perform TF-IDF search
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

// Perform secondary letter-by-letter search for suggestions
func letterByLetterSearch(query string) []int {
	query = strings.ToLower(query)
	result := make(map[int]bool)

	for _, doc := range docs {
		text := strings.ToLower(doc.Text)
		if strings.Contains(text, query) {
			result[doc.ID] = true
		}
	}

	var docIDs []int
	for docID := range result {
		docIDs = append(docIDs, docID)
	}

	return docIDs
}

// Rank search results
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

// Serve the index.html page
func serveIndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

// Handle search requests
func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	searchResults := searchTFIDF(tfidfIndex, query)
	rankedResults := rankSearchResults(searchResults)

	// Perform letter-by-letter search for suggestions
	letterSearchResults := letterByLetterSearch(query)
	suggestions := make(map[int]bool)
	for _, docID := range letterSearchResults {
		suggestions[docID] = true
	}

	// Combine ranked results and suggestions
	var results []string
	for _, docID := range rankedResults {
		if _, ok := suggestions[docID]; ok {
			for _, doc := range docs {
				if doc.ID == docID {
					results = append(results, doc.Text)
					delete(suggestions, docID)
					break
				}
			}
		}
	}
	for docID := range suggestions {
		for _, doc := range docs {
			if doc.ID == docID {
				results = append(results, doc.Text)
				break
			}
		}
	}

	jsonResponse, err := json.Marshal(results)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func main() {
	http.HandleFunc("/", serveIndexPage)
	http.HandleFunc("/search", handleSearch)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
