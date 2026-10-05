// Package domain contains the core domain models for knowledge retrieval.
package domain

import "time"

// PolicyChunk represents a chunk of text extracted from a policy document.
type PolicyChunk struct {
	ID        string
	Source    string
	Heading   string
	Text      string
	Language  string
	Version   string
	Vector    []float64
	Metadata  map[string]string
	CreatedAt time.Time
}

// DocumentMeta represents metadata about a knowledge document.
type DocumentMeta struct {
	ID     string
	Source string
	Title  string
}

// RetrievalResult represents a result from a knowledge retrieval query.
type RetrievalResult struct {
	Chunk PolicyChunk
	Score float64
}
