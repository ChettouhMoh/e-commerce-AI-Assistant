package application

import (
	"context"

	"ecommerce-ai-assistant/internal/knowledge/domain"
	"ecommerce-ai-assistant/internal/knowledge/ports"
)

type KnowledgeService struct {
	repo  ports.KnowledgeRepository
	embed ports.EmbeddingProvider
	topK  int
}

func NewKnowledgeService(repo ports.KnowledgeRepository, embed ports.EmbeddingProvider, topK int) *KnowledgeService {
	return &KnowledgeService{repo: repo, embed: embed, topK: topK}
}

func (s *KnowledgeService) Search(ctx context.Context, query string) ([]domain.RetrievalResult, error) {
	if s.embed == nil {
		return nil, nil
	}
	return s.repo.Search(ctx, query, s.topK)
}
