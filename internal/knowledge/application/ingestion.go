package application

import (
	"bufio"
	"context"
	"io"
	"regexp"
	"strings"

	"ecommerce-ai-assistant/internal/knowledge/domain"
)

var headingRE = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

type Chunk struct {
	Text     string
	Heading  string
	Source   string
	Language string
	Version  string
}

type Chunker struct {
	chunkSize    int
	chunkOverlap int
}

func NewChunker(chunkSize, chunkOverlap int) *Chunker {
	if chunkOverlap >= chunkSize {
		chunkOverlap = chunkSize / 4
	}
	return &Chunker{chunkSize: chunkSize, chunkOverlap: chunkOverlap}
}

func (c *Chunker) ChunkDocuments(ctx context.Context, docs []domain.DocumentMeta, readerFn func(domain.DocumentMeta) (io.Reader, error)) ([]Chunk, error) {
	var allChunks []Chunk

	for _, doc := range docs {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		r, err := readerFn(doc)
		if err != nil {
			return nil, err
		}

		chunks, err := c.chunkReader(r, doc)
		if err != nil {
			return nil, err
		}
		allChunks = append(allChunks, chunks...)
	}

	return allChunks, nil
}

func (c *Chunker) chunkReader(r io.Reader, meta domain.DocumentMeta) ([]Chunk, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentHeading string
	var paragraphs []string
	var chunks []Chunk

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if m := headingRE.FindStringSubmatch(trimmed); m != nil {
			if len(paragraphs) > 0 {
				chunks = append(chunks, c.flush(paragraphs, currentHeading, meta)...)
				paragraphs = nil
			}
			currentHeading = m[2]
			continue
		}

		if trimmed == "" {
			if len(paragraphs) > 0 {
				chunks = append(chunks, c.flush(paragraphs, currentHeading, meta)...)
				paragraphs = nil
			}
			continue
		}

		paragraphs = append(paragraphs, trimmed)
	}

	if len(paragraphs) > 0 {
		chunks = append(chunks, c.flush(paragraphs, currentHeading, meta)...)
	}

	return chunks, scanner.Err()
}

func (c *Chunker) flush(paragraphs []string, heading string, meta domain.DocumentMeta) []Chunk {
	var result []Chunk
	text := strings.Join(paragraphs, " ")
	heading = strings.TrimSpace(heading)
	if heading == "" {
		heading = meta.Title
	}
	source := meta.Source
	if source == "" {
		source = meta.ID
	}

	for len(text) > c.chunkSize {
		cut := c.chunkSize
		if cut < len(text) {
			for i := cut; i > cut-c.chunkOverlap && i > 0; i-- {
				if text[i] == ' ' {
					cut = i
					break
				}
			}
		}
		result = append(result, Chunk{
			Text:    strings.TrimSpace(text[:cut]),
			Heading: heading,
			Source:  source,
		})
		text = strings.TrimSpace(text[cut:])
	}

	if text != "" {
		result = append(result, Chunk{
			Text:    text,
			Heading: heading,
			Source:  source,
		})
	}

	return result
}
