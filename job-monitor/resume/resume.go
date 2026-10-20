package resume

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pdfapi "github.com/pdfcpu/pdfcpu/pkg/api"
)

func LoadText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return extractPDF(path)
	case ".txt", ".md":
		return readFile(path)
	default:
		return "", fmt.Errorf("unsupported resume format %q (use .pdf, .txt, or .md)", ext)
	}
}

func extractPDF(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	err = pdfapi.ExtractContent(context.Background(), f, nil, func(r io.Reader, pageNr int) error {
		_, err := io.Copy(&buf, r)
		return err
	}, nil)
	if err != nil {
		return "", fmt.Errorf("extract pdf text: %w", err)
	}
	return strings.TrimSpace(buf.String()), nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
