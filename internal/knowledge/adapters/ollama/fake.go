package ollama

import (
	"context"
	"math"
	"sync"
)

type FakeEmbedding struct {
	mu      sync.Mutex
	vectors map[string][]float64
	dim     int
}

func NewFakeEmbedding(dim int) *FakeEmbedding {
	return &FakeEmbedding{
		vectors: make(map[string][]float64),
		dim:     dim,
	}
}

func (f *FakeEmbedding) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	results := make([][]float64, len(texts))
	for i, text := range texts {
		results[i] = f.embedOne(text)
	}
	return results, nil
}

func (f *FakeEmbedding) embedOne(text string) []float64 {
	f.mu.Lock()
	defer f.mu.Unlock()

	if v, ok := f.vectors[text]; ok {
		return v
	}

	v := hashVector(text, f.dim)
	f.vectors[text] = v
	return v
}

func hashVector(s string, dim int) []float64 {
	v := make([]float64, dim)
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	for i := 0; i < dim; i++ {
		h = h*31 + uint32(i)
		v[i] = float64(int32(h%1000)) / 500.0
	}
	return normalize(v)
}

func normalize(v []float64) []float64 {
	var sum float64
	for _, f := range v {
		sum += f * f
	}
	if sum == 0 {
		return v
	}
	mag := math.Sqrt(sum)
	out := make([]float64, len(v))
	for i, f := range v {
		out[i] = f / mag
	}
	return out
}

func (f *FakeEmbedding) SetCached(text string, vector []float64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.vectors[text] = vector
}

func (f *FakeEmbedding) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.vectors = make(map[string][]float64)
}

func (f *FakeEmbedding) Dimension() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.dim
}
