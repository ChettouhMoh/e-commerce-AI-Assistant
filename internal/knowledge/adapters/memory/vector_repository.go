package memory

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	domain "ecommerce-ai-assistant/internal/knowledge/domain"
	"ecommerce-ai-assistant/internal/knowledge/ports"
)

type VectorRepository struct {
	chunks []domain.PolicyChunk
	mu     sync.RWMutex
}

var _ ports.KnowledgeRepository = (*VectorRepository)(nil)

func NewVectorRepository() *VectorRepository {
	r := &VectorRepository{
		chunks: make([]domain.PolicyChunk, 0),
	}
	r.Seed()
	return r
}

func NewEmptyVectorRepository() *VectorRepository {
	return &VectorRepository{
		chunks: make([]domain.PolicyChunk, 0),
	}
}

func (r *VectorRepository) Search(ctx context.Context, query string, topK int) ([]domain.RetrievalResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.chunks) == 0 || topK <= 0 {
		return nil, nil
	}

	queryVec := deterministicEmbedding(len(r.chunks[0].Vector), 0.92)

	type scored struct {
		idx   int
		score float64
	}
	scores := make([]scored, 0, len(r.chunks))
	for i, c := range r.chunks {
		if len(c.Vector) != len(queryVec) || isZeroVector(c.Vector) {
			continue
		}
		score := cosineSimilarity(queryVec, c.Vector)
		scores = append(scores, scored{idx: i, score: score})
	}

	sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })
	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]domain.RetrievalResult, 0, topK)
	for i := 0; i < topK; i++ {
		results = append(results, domain.RetrievalResult{
			Chunk: r.chunks[scores[i].idx],
			Score: scores[i].score,
		})
	}
	return results, nil
}

func (r *VectorRepository) Store(ctx context.Context, chunks []domain.PolicyChunk) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range chunks {
		if chunks[i].ID == "" {
			chunks[i].ID = "CHUNK-" + time.Now().Format("20060102150405")
		}
		if chunks[i].Source == "" {
			chunks[i].Source = "unknown"
		}
		if chunks[i].Text == "" {
			continue
		}
		if chunks[i].Vector == nil {
			chunks[i].Vector = deterministicEmbedding(384, 0.80)
		}
		if chunks[i].Metadata == nil {
			chunks[i].Metadata = make(map[string]string)
		}
		r.chunks = append(r.chunks, chunks[i])
	}
	return nil
}

func cosineSimilarity(a, b []float64) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func isZeroVector(v []float64) bool {
	for _, f := range v {
		if f != 0 {
			return false
		}
	}
	return true
}

func deterministicEmbedding(dim int, base float64) []float64 {
	vec := make([]float64, dim)
	for i := range vec {
		vec[i] = base + float64(i%10)*0.001
	}
	return vec
}

func (r *VectorRepository) Seed() {
	chunks := []domain.PolicyChunk{
		{ID: "CHUNK-001", Source: "policies/returns.md", Text: "Customers may return unworn items within 30 days of delivery with original tags attached.", Vector: deterministicEmbedding(384, 0.92), Metadata: map[string]string{"section": "returns"}},
		{ID: "CHUNK-002", Source: "policies/returns.md", Text: "Refunds are processed within 5 to 7 business days after the returned item is received and inspected.", Vector: deterministicEmbedding(384, 0.88), Metadata: map[string]string{"section": "refunds"}},
		{ID: "CHUNK-003", Source: "policies/returns.md", Text: "Sale items are final sale and cannot be returned unless defective.", Vector: deterministicEmbedding(384, 0.85), Metadata: map[string]string{"section": "returns"}},
		{ID: "CHUNK-004", Source: "policies/shipping.md", Text: "Standard shipping typically takes 5 to 7 business days. Express shipping takes 2 to 3 business days.", Vector: deterministicEmbedding(384, 0.90), Metadata: map[string]string{"section": "shipping"}},
		{ID: "CHUNK-005", Source: "policies/shipping.md", Text: "Free shipping is available on all domestic orders over 75 USD.", Vector: deterministicEmbedding(384, 0.86), Metadata: map[string]string{"section": "shipping"}},
		{ID: "CHUNK-006", Source: "policies/shipping.md", Text: "International shipping is available to select countries with additional customs fees.", Vector: deterministicEmbedding(384, 0.82), Metadata: map[string]string{"section": "shipping"}},
		{ID: "CHUNK-007", Source: "policies/exchanges.md", Text: "Exchanges for different sizes or colors are free within 60 days of purchase.", Vector: deterministicEmbedding(384, 0.94), Metadata: map[string]string{"section": "exchanges"}},
		{ID: "CHUNK-008", Source: "policies/exchanges.md", Text: "If the requested exchange item is out of stock, a store credit will be issued.", Vector: deterministicEmbedding(384, 0.90), Metadata: map[string]string{"section": "exchanges"}},
		{ID: "CHUNK-009", Source: "policies/warranty.md", Text: "All electronics carry a 12 month limited warranty against manufacturing defects.", Vector: deterministicEmbedding(384, 0.88), Metadata: map[string]string{"section": "warranty"}},
		{ID: "CHUNK-010", Source: "policies/warranty.md", Text: "Warranty claims must be submitted within 14 days of discovering the defect.", Vector: deterministicEmbedding(384, 0.85), Metadata: map[string]string{"section": "warranty"}},
		{ID: "CHUNK-011", Source: "policies/payment.md", Text: "We accept Visa, Mastercard, American Express, and PayPal.", Vector: deterministicEmbedding(384, 0.82), Metadata: map[string]string{"section": "payment"}},
		{ID: "CHUNK-012", Source: "policies/payment.md", Text: "Buy now pay later is available through Affirm on orders over 50 USD.", Vector: deterministicEmbedding(384, 0.80), Metadata: map[string]string{"section": "payment"}},
	}
	r.chunks = chunks
}
