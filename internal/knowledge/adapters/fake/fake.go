package fake

import (
	"context"
)

type FakeEmbeddingProvider struct {
	vectors map[string][]float64
	dim     int
}

func NewFakeEmbeddingProvider() *FakeEmbeddingProvider {
	return &FakeEmbeddingProvider{
		vectors: make(map[string][]float64),
		dim:     384,
	}
}

func (f *FakeEmbeddingProvider) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	result := make([][]float64, len(texts))
	for i, text := range texts {
		if v, ok := f.vectors[text]; ok {
			result[i] = v
		} else {
			result[i] = make([]float64, f.dim)
			for j := range result[i] {
				result[i][j] = float64(len(text)) / 100.0
			}
		}
	}
	return result, nil
}

func (f *FakeEmbeddingProvider) Dimension() int {
	return f.dim
}

func (f *FakeEmbeddingProvider) SetVector(text string, vec []float64) {
	f.vectors[text] = vec
}
