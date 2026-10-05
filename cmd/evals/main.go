package main

import (
	"context"
	"fmt"
	"log"

	evaluation "ecommerce-ai-assistant/evals"
	"ecommerce-ai-assistant/internal/knowledge/adapters/memory"
)

func main() {
	repo := memory.NewVectorRepository()
	ctx := context.Background()

	cases := []evaluation.RetrievalCase{
		{Question: "return policy", Expected: []string{"CHUNK-001", "CHUNK-002"}},
	}

	results, err := evaluation.EvaluateRetrieval(ctx, repo, cases)
	if err != nil {
		log.Fatalf("evaluation failed: %v", err)
	}

	metrics := evaluation.AggregateMetrics(results)
	fmt.Printf("Recall@10: %.2f, MRR: %.2f\n", metrics["recall@10"], metrics["mrr"])
}
