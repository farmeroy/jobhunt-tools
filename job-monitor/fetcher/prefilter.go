package fetcher

import (
	"regexp"
	"strings"
)

var titleWhitelist = []string{
	"engineer",
	"developer",
	"software",
	"sre",
	"devops",
	"platform",
	"infrastructure",
	"backend",
	"frontend",
	"fullstack",
	"full-stack",
	"full stack",
}

var locationPatterns = []string{
	"remote",
	"united states",
	"hybrid",
	"california",
	"san francisco",
	"new york",
	"arizona",
	"texas",
	"colorado",
	"washington",
	"oregon",
	"virginia",
	"georgia",
	"illinois",
	"massachusetts",
	"pennsylvania",
	"north carolina",
}

// Matches "US" as a whole word — not inside "Marcus", "Focus", etc.
var usWordBoundary = regexp.MustCompile(`(?i)\bUS\b`)

func MatchesTitle(title string) bool {
	lower := strings.ToLower(title)
	for _, kw := range titleWhitelist {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func MatchesLocation(location string) bool {
	if location == "" {
		return true
	}
	lower := strings.ToLower(location)
	for _, kw := range locationPatterns {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return usWordBoundary.MatchString(location)
}

func PreFilter(postings []JobPosting) []JobPosting {
	var out []JobPosting
	for _, p := range postings {
		if MatchesTitle(p.Title) && MatchesLocation(p.Location) {
			out = append(out, p)
		}
	}
	return out
}
