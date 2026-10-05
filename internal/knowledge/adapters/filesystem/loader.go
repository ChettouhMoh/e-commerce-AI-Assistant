package filesystem

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DocumentLoader struct {
	rootDir string
}

func NewDocumentLoader(rootDir string) *DocumentLoader {
	return &DocumentLoader{rootDir: rootDir}
}

func (d *DocumentLoader) Load(ctx context.Context, source string) (io.Reader, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("empty source path")
	}

	path := filepath.Join(d.rootDir, source)
	path = filepath.Clean(path)

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", source, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("source %s is a directory, expected file", source)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", source, err)
	}

	return &fsReader{File: f}, nil
}

func (d *DocumentLoader) Discover(ctx context.Context, exts ...string) ([]string, error) {
	var files []string
	extSet := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		extSet["."+strings.TrimPrefix(e, ".")] = struct{}{}
	}

	err := filepath.Walk(d.rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() {
			return nil
		}
		if len(extSet) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			if _, ok := extSet[ext]; !ok {
				return nil
			}
		}
		rel, err := filepath.Rel(d.rootDir, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})

	return files, err
}

type fsReader struct {
	*os.File
}

func (f *fsReader) Close() error {
	return f.File.Close()
}
