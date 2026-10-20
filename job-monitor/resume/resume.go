package resume

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	out, err := exec.Command("pdftotext", "-layout", path, "-").Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext: %w (is poppler-utils installed?)", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
