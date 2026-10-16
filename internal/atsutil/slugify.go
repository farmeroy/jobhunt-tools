package atsutil

import (
	"strings"
	"unicode"
)

func Slugify(company string) []string {
	lower := strings.ToLower(company)

	clean := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' {
			return r
		}
		return -1
	}, lower)

	fields := strings.Fields(clean)
	joined := strings.Join(fields, "")
	dashed := strings.Join(fields, "-")

	slugs := []string{joined}
	if dashed != joined {
		slugs = append(slugs, dashed)
	}

	for _, suffix := range []string{"ai", "io", "co", "inc", "health"} {
		for _, s := range []string{joined, dashed} {
			trimmed := strings.TrimSuffix(s, suffix)
			trimmed = strings.TrimSuffix(trimmed, "-")
			if trimmed != "" && trimmed != s {
				slugs = append(slugs, trimmed)
			}
		}
	}

	seen := map[string]bool{}
	var unique []string
	for _, s := range slugs {
		if !seen[s] {
			seen[s] = true
			unique = append(unique, s)
		}
	}
	return unique
}
