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

	"github.com/PuerkitoBio/goquery"
	"github.com/ledongthuc/pdf"
)

type Document interface {
	GetID() int
	GetText() string
}

type GenericDocument struct {
	ID   int
	Text string
}

func (d GenericDocument) GetID() int {
	return d.ID
}

func (d GenericDocument) GetText() string {
	return d.Text
}

type WebDocument struct {
	ID  int
	URL string
}

func (d WebDocument) GetID() int {
	return d.ID
}

func (d WebDocument) GetText() string {
	resp, err := http.Get(d.URL)
	if err != nil {
		log.Printf("Error fetching URL %s: %v", d.URL, err)
		return ""
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("Error parsing HTML from %s: %v", d.URL, err)
		return ""
	}

	return doc.Find("body").Text()
}

type PDFDocument struct {
	ID   int
	Path string
}

func (d PDFDocument) GetID() int {
	return d.ID
}

func (d PDFDocument) GetText() string {
	f, r, err := pdf.Open(d.Path)
	if err != nil {
		log.Printf("Error opening PDF %s: %v", d.Path, err)
		return ""
	}
	defer f.Close()

	var text string
	for pageIndex := 1; pageIndex <= r.NumPage(); pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}
		content, _ := p.GetPlainText(nil)
		text += content
	}
	return text
}

type TermFrequency map[string]float64
type InvertedIndex map[string][]int
type TFIDFIndex map[string]map[int]float64

type FTS struct {
	Documents     []Document
	InvertedIndex InvertedIndex
	TFIDFIndex    TFIDFIndex
}

func NewFTS() *FTS {
	return &FTS{
		Documents:     []Document{},
		InvertedIndex: make(InvertedIndex),
		TFIDFIndex:    make(TFIDFIndex),
	}
}

func (fts *FTS) AddDocument(doc Document) {
	fts.Documents = append(fts.Documents, doc)
}

func (fts *FTS) Start() {
	fts.buildInvertedIndex()
	fts.buildTFIDFIndex()
}

func (fts *FTS) buildInvertedIndex() {
	for _, doc := range fts.Documents {
		tokens := tokenize(doc.GetText())
		for _, token := range tokens {
			fts.InvertedIndex[token] = append(fts.InvertedIndex[token], doc.GetID())
		}
	}
}

func (fts *FTS) buildTFIDFIndex() {
	totalDocs := len(fts.Documents)
	idf := calculateIDF(fts.InvertedIndex, totalDocs)

	for _, doc := range fts.Documents {
		tokens := tokenize(doc.GetText())
		docTF := calculateTermFrequency(tokens)
		for token, tf := range docTF {
			idfValue := idf[token]
			if _, ok := fts.TFIDFIndex[token]; !ok {
				fts.TFIDFIndex[token] = make(map[int]float64)
			}
			fts.TFIDFIndex[token][doc.GetID()] = tf * idfValue
		}
	}
}

func (fts *FTS) Search(query string) map[int]float64 {
	terms := tokenize(query)
	queryTF := calculateTermFrequency(terms)

	result := make(map[int]float64)

	for term, tf := range queryTF {
		if docIDs, ok := fts.TFIDFIndex[term]; ok {
			for docID, tfidf := range docIDs {
				result[docID] += tf * tfidf
			}
		}
	}

	return result
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

// Calculate IDF
func calculateIDF(index InvertedIndex, totalDocs int) map[string]float64 {
	idf := make(map[string]float64)
	for term, docIDs := range index {
		idf[term] = math.Log(float64(totalDocs) / float64(len(docIDs)))
	}
	return idf
}

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

func serveIndexPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func handleSearch(w http.ResponseWriter, r *http.Request, fts *FTS) {
	query := r.URL.Query().Get("query")
	searchResults := fts.Search(query)
	rankedResults := rankSearchResults(searchResults)

	var results []string
	for _, docID := range rankedResults {
		for _, doc := range fts.Documents {
			if doc.GetID() == docID {
				results = append(results, doc.GetText())
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
	fts := NewFTS()

	// Add generic documents
	fts.AddDocument(GenericDocument{ID: 1, Text: "This is a document about Go programming."})
	fts.AddDocument(GenericDocument{ID: 2, Text: "Go is a statically typed, compiled programming language."})

	// Add a web document
	fts.AddDocument(WebDocument{ID: 3, URL: "https://learnopencv.com/handwritten-text-recognition-using-ocr/"})

	// Add a PDF document
	fts.AddDocument(PDFDocument{ID: 4, Path: "document.pdf"})

	fts.Start()

	http.HandleFunc("/", serveIndexPage)
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		handleSearch(w, r, fts)
	})

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
