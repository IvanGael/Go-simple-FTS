package main

import (
	"reflect"
	"testing"
)

func TestGenericDocument(t *testing.T) {
	doc := GenericDocument{ID: 1, Text: "Test document"}
	if doc.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", doc.GetID())
	}
	if doc.GetText() != "Test document" {
		t.Errorf("Expected text 'Test document', got '%s'", doc.GetText())
	}
}

func TestWebDocument(t *testing.T) {
	doc := WebDocument{ID: 1, URL: "https://learnopencv.com/handwritten-text-recognition-using-ocr/"}
	if doc.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", doc.GetID())
	}
	if doc.GetText() == "" {
		t.Errorf("Expected text , got '%s'", doc.GetText())
	}
}

func TestPDFDocument(t *testing.T) {
	doc := PDFDocument{ID: 1, Path: "document.pdf"}
	if doc.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", doc.GetID())
	}
	if doc.GetText() == "" {
		t.Errorf("Expected text , got '%s'", doc.GetText())
	}
}

func TestTokenize(t *testing.T) {
	text := "This is a TEST document"
	expected := []string{"this", "is", "a", "test", "document"}
	result := tokenize(text)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCalculateTermFrequency(t *testing.T) {
	tokens := []string{"this", "is", "a", "test", "document", "this", "is", "a", "test"}
	expected := TermFrequency{
		"this":     2.0 / 9,
		"is":       2.0 / 9,
		"a":        2.0 / 9,
		"test":     2.0 / 9,
		"document": 1.0 / 9,
	}
	result := calculateTermFrequency(tokens)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestCalculateIDF(t *testing.T) {
	index := InvertedIndex{
		"test":     []int{1, 2, 3},
		"document": []int{1, 2},
		"rare":     []int{3, 5},
	}
	totalDocs := 3
	expected := map[string]float64{
		"test":     0,
		"document": 0.4054651081081644,
		"rare":     0.4054651081081644,
	}
	result := calculateIDF(index, totalDocs)
	for term, expectedValue := range expected {
		if result[term] != expectedValue {
			t.Errorf("For term '%s', expected %f, got %f", term, expectedValue, result[term])
		}
	}
}

func TestNewFTS(t *testing.T) {
	fts := NewFTS()
	if fts.Documents == nil {
		t.Error("Documents slice should be initialized")
	}
	if fts.InvertedIndex == nil {
		t.Error("InvertedIndex should be initialized")
	}
	if fts.TFIDFIndex == nil {
		t.Error("TFIDFIndex should be initialized")
	}
}

func TestAddDocument(t *testing.T) {
	fts := NewFTS()
	doc := GenericDocument{ID: 1, Text: "Test document"}
	fts.AddDocument(doc)
	if len(fts.Documents) != 1 {
		t.Errorf("Expected 1 document, got %d", len(fts.Documents))
	}
	if fts.Documents[0] != doc {
		t.Error("Added document does not match the original")
	}
}

func TestBuildInvertedIndex(t *testing.T) {
	fts := NewFTS()
	fts.AddDocument(GenericDocument{ID: 1, Text: "This is a test"})
	fts.AddDocument(GenericDocument{ID: 2, Text: "This is another test"})
	fts.buildInvertedIndex()

	expected := InvertedIndex{
		"this":    []int{1, 2},
		"is":      []int{1, 2},
		"a":       []int{1},
		"test":    []int{1, 2},
		"another": []int{2},
	}

	if !reflect.DeepEqual(fts.InvertedIndex, expected) {
		t.Errorf("Expected %v, got %v", expected, fts.InvertedIndex)
	}
}

func TestSearch(t *testing.T) {
	fts := NewFTS()
	fts.AddDocument(GenericDocument{ID: 1, Text: "This is a test document"})
	fts.AddDocument(GenericDocument{ID: 2, Text: "This is another document"})
	fts.Start()

	results := fts.Search("test document")
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
	if results[1] <= results[2] {
		t.Error("Document 1 should have higher score than document 2")
	}
}

func TestRankSearchResults(t *testing.T) {
	results := map[int]float64{
		1: 0.5,
		2: 0.8,
		3: 0.2,
	}
	expected := []int{2, 1, 3}
	ranked := rankSearchResults(results)
	if !reflect.DeepEqual(ranked, expected) {
		t.Errorf("Expected %v, got %v", expected, ranked)
	}
}
