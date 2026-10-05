package evals

import (
	"context"

	"ecommerce-ai-assistant/internal/knowledge/adapters/memory"
)

type RetrievalCase struct {
	Question string   `json:"question"`
	Expected []string `json:"expected_chunk_ids"`
}

type RetrievalResult struct {
	Case   RetrievalCase
	Found  []string
	Recall float64
	MRR    float64
}

func EvaluateRetrieval(ctx context.Context, repo *memory.VectorRepository, cases []RetrievalCase) ([]RetrievalResult, error) {
	results := make([]RetrievalResult, 0, len(cases))
	for _, c := range cases {
		found := make([]string, 0)
		hits, err := repo.Search(ctx, c.Question, 10)
		if err != nil {
			return nil, err
		}
		for _, h := range hits {
			found = append(found, h.Chunk.ID)
		}

		recall := recallAtK(found, c.Expected, 10)
		mrr := reciprocalRank(found, c.Expected)

		results = append(results, RetrievalResult{
			Case:   c,
			Found:  found,
			Recall: recall,
			MRR:    mrr,
		})
	}
	return results, nil
}

func recallAtK(found, expected []string, k int) float64 {
	if len(expected) == 0 {
		return 0
	}
	expectedSet := make(map[string]bool)
	for _, e := range expected {
		expectedSet[e] = true
	}
	hits := 0
	for i, f := range found {
		if i >= k {
			break
		}
		if expectedSet[f] {
			hits++
		}
	}
	return float64(hits) / float64(len(expected))
}

func reciprocalRank(found, expected []string) float64 {
	expectedSet := make(map[string]bool)
	for _, e := range expected {
		expectedSet[e] = true
	}
	for i, f := range found {
		if expectedSet[f] {
			return 1.0 / float64(i+1)
		}
	}
	return 0
}

func AggregateMetrics(results []RetrievalResult) map[string]float64 {
	var totalRecall, totalMRR float64
	for _, r := range results {
		totalRecall += r.Recall
		totalMRR += r.MRR
	}
	return map[string]float64{
		"recall@10": totalRecall / float64(len(results)),
		"mrr":       totalMRR / float64(len(results)),
	}
}
