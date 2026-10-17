package fetcher

import "testing"

func TestMatchesTitle(t *testing.T) {
	tests := []struct {
		title string
		want  bool
	}{
		{"Software Engineer", true},
		{"Senior Backend Developer", true},
		{"SRE Team Lead", true},
		{"DevOps Engineer", true},
		{"Platform Engineer", true},
		{"Infrastructure Engineer", true},
		{"Frontend Developer", true},
		{"Full-Stack Engineer", true},
		{"Full Stack Developer", true},
		{"Fullstack Engineer", true},
		{"Software Analyst", true},
		{"Account Manager", false},
		{"Product Designer", false},
		{"Marketing Lead", false},
		{"VP of Sales", false},
		{"Recruiter", false},
	}
	for _, tt := range tests {
		if got := MatchesTitle(tt.title); got != tt.want {
			t.Errorf("MatchesTitle(%q) = %v, want %v", tt.title, got, tt.want)
		}
	}
}

func TestMatchesLocation(t *testing.T) {
	tests := []struct {
		location string
		want     bool
	}{
		{"", true},
		{"Remote", true},
		{"Remote - US", true},
		{"San Francisco, California, United States", true},
		{"New York, United States", true},
		{"Hybrid", true},
		{"NerdWallet US", true},
		{"Scottsdale, Arizona", true},
		{"London, United Kingdom", false},
		{"Bangalore, India", false},
		{"Berlin, Germany", false},
		{"NerdWallet Canada", false},
		{"Toronto, Canada", false},
		{"Focus Labs", false},
	}
	for _, tt := range tests {
		if got := MatchesLocation(tt.location); got != tt.want {
			t.Errorf("MatchesLocation(%q) = %v, want %v", tt.location, got, tt.want)
		}
	}
}

func TestPreFilter(t *testing.T) {
	postings := []JobPosting{
		{Title: "Software Engineer", Location: "Remote - US"},
		{Title: "Account Manager", Location: "Remote"},
		{Title: "Backend Developer", Location: "London, UK"},
		{Title: "Frontend Developer", Location: "San Francisco, California, United States"},
		{Title: "Marketing Lead", Location: "Berlin, Germany"},
		{Title: "Staff Software Engineer", Location: "NerdWallet US"},
		{Title: "Software Engineer", Location: "NerdWallet Canada"},
	}

	got := PreFilter(postings)
	if len(got) != 3 {
		t.Fatalf("expected 3 postings after filter, got %d", len(got))
	}
	if got[0].Title != "Software Engineer" {
		t.Errorf("expected Software Engineer, got %s", got[0].Title)
	}
	if got[1].Title != "Frontend Developer" {
		t.Errorf("expected Frontend Developer, got %s", got[1].Title)
	}
	if got[2].Title != "Staff Software Engineer" {
		t.Errorf("expected Staff Software Engineer, got %s", got[2].Title)
	}
}
