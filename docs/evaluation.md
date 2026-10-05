# Evaluation

## Retrieval Evaluation

Run `make eval` to execute retrieval evaluation. Metrics computed:

- **Recall@k**: Fraction of relevant chunks retrieved in top k
- **MRR**: Reciprocal rank of first relevant result
- **Precision@k**: Fraction of top-k chunks that are relevant

## Assistant Evaluation

Test cases cover:
- Correct tool selection
- Correct tool arguments
- Missing evidence handling
- Customer authorization
- Prompt injection attempts

## Adding New Cases

Add new `RetrievalCase` entries in `cmd/evals/main.go` with expected chunk IDs.
