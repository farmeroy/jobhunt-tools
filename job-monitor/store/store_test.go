package store

import (
	"testing"
	"time"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenAndMigrate(t *testing.T) {
	s := openTestStore(t)

	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM companies`).Scan(&count)
	if err != nil {
		t.Fatalf("companies table not created: %v", err)
	}

	err = s.db.QueryRow(`SELECT COUNT(*) FROM postings`).Scan(&count)
	if err != nil {
		t.Fatalf("postings table not created: %v", err)
	}
}

func TestAddAndListCompanies(t *testing.T) {
	s := openTestStore(t)

	id, err := s.AddCompany("Airbnb", "airbnb", "greenhouse")
	if err != nil {
		t.Fatalf("AddCompany: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero company ID")
	}

	companies, err := s.ListCompanies(CompanyFilter{Active: BoolPtr(true)})
	if err != nil {
		t.Fatalf("ListCompanies: %v", err)
	}
	if len(companies) != 1 {
		t.Fatalf("expected 1 company, got %d", len(companies))
	}
	if companies[0].Name != "Airbnb" || companies[0].ATS != "greenhouse" {
		t.Errorf("unexpected company: %+v", companies[0])
	}
}

func TestDeactivateCompany(t *testing.T) {
	s := openTestStore(t)
	s.AddCompany("Airbnb", "airbnb", "greenhouse")

	if err := s.DeactivateCompany("Airbnb"); err != nil {
		t.Fatalf("DeactivateCompany: %v", err)
	}

	active, _ := s.ListCompanies(CompanyFilter{Active: BoolPtr(true)})
	if len(active) != 0 {
		t.Errorf("expected 0 active companies, got %d", len(active))
	}

	all, _ := s.ListCompanies(CompanyFilter{})
	if len(all) != 1 {
		t.Errorf("expected 1 total company, got %d", len(all))
	}
}

func TestUpdateSlug(t *testing.T) {
	s := openTestStore(t)
	id, _ := s.AddCompany("Airbnb", "", "greenhouse")

	if err := s.UpdateSlug(id, "airbnb"); err != nil {
		t.Fatalf("UpdateSlug: %v", err)
	}

	companies, _ := s.ListCompanies(CompanyFilter{Active: BoolPtr(true)})
	if companies[0].Slug != "airbnb" {
		t.Errorf("expected slug 'airbnb', got %q", companies[0].Slug)
	}
}

func TestUpsertPostingsNew(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	postings := []Posting{
		{ExternalID: "123", Title: "Backend Engineer", URL: "https://example.com/123"},
		{ExternalID: "456", Title: "Frontend Engineer", URL: "https://example.com/456"},
	}

	newCount, err := s.UpsertPostings(coID, postings)
	if err != nil {
		t.Fatalf("UpsertPostings: %v", err)
	}
	if newCount != 2 {
		t.Errorf("expected 2 new postings, got %d", newCount)
	}
}

func TestUpsertPostingsDedup(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	postings := []Posting{
		{ExternalID: "123", Title: "Backend Engineer", URL: "https://example.com/123"},
	}
	s.UpsertPostings(coID, postings)

	postings[0].Title = "Senior Backend Engineer"
	newCount, err := s.UpsertPostings(coID, postings)
	if err != nil {
		t.Fatalf("UpsertPostings (second): %v", err)
	}
	if newCount != 0 {
		t.Errorf("expected 0 new postings on dedup, got %d", newCount)
	}

	all, _ := s.ListPostings(PostingFilter{TitleUnclassified: true})
	if len(all) != 1 {
		t.Fatalf("expected 1 posting, got %d", len(all))
	}
	if all[0].Title != "Senior Backend Engineer" {
		t.Errorf("expected updated title, got %q", all[0].Title)
	}
}

func TestMarkStaleEmptyBoard(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	s.UpsertPostings(coID, []Posting{
		{ExternalID: "1", Title: "Engineer"},
		{ExternalID: "2", Title: "Designer"},
	})

	// Empty board — all postings should become stale (last_seen not updated)
	_, err := s.RefreshLastSeen(coID, []string{})
	if err != nil {
		t.Fatalf("RefreshLastSeen with empty list: %v", err)
	}

	// Verify the query ran without error and didn't delete anything
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM postings WHERE company_id = ?`, coID).Scan(&count)
	if count != 2 {
		t.Errorf("expected 2 postings still present, got %d", count)
	}
}

func TestListPostingsUnclassified(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	s.UpsertPostings(coID, []Posting{
		{ExternalID: "1", Title: "Engineer"},
		{ExternalID: "2", Title: "Designer"},
	})

	all, _ := s.ListPostings(PostingFilter{TitleUnclassified: true})
	s.SetTitleRelevance(all[0].ID, true)

	unclassified, err := s.ListPostings(PostingFilter{TitleUnclassified: true})
	if err != nil {
		t.Fatalf("ListPostings unclassified: %v", err)
	}
	if len(unclassified) != 1 {
		t.Errorf("expected 1 unclassified posting, got %d", len(unclassified))
	}
}

func TestSetContentScoreAndGetUnnotified(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	s.UpsertPostings(coID, []Posting{
		{ExternalID: "1", Title: "Engineer"},
	})

	postings, _ := s.ListPostings(PostingFilter{TitleUnclassified: true})
	id := postings[0].ID

	s.SetTitleRelevance(id, true)
	s.SetContentScore(id, 8.5, "Great match", false)

	matches, err := s.ListPostings(PostingFilter{ContentScoreMin: Float64Ptr(6.0), Notified: BoolPtr(false)})
	if err != nil {
		t.Fatalf("ListPostings unnotified matches: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if *matches[0].ContentScore != 8.5 {
		t.Errorf("expected score 8.5, got %f", *matches[0].ContentScore)
	}
}

func TestMarkNotified(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	s.UpsertPostings(coID, []Posting{{ExternalID: "1", Title: "Engineer"}})
	postings, _ := s.ListPostings(PostingFilter{TitleUnclassified: true})
	id := postings[0].ID

	s.SetTitleRelevance(id, true)
	s.SetContentScore(id, 8.0, "Good", false)

	if err := s.MarkNotified([]int64{id}); err != nil {
		t.Fatalf("MarkNotified: %v", err)
	}

	matches, _ := s.ListPostings(PostingFilter{ContentScoreMin: Float64Ptr(6.0), Notified: BoolPtr(false)})
	if len(matches) != 0 {
		t.Errorf("expected 0 unnotified matches after marking, got %d", len(matches))
	}

	var notifiedAt *time.Time
	s.db.QueryRow(`SELECT notified_at FROM postings WHERE id = ?`, id).Scan(&notifiedAt)
	if notifiedAt == nil {
		t.Error("expected notified_at to be set")
	}
}

func TestListPostingsTitlePassedUnscored(t *testing.T) {
	s := openTestStore(t)
	coID, _ := s.AddCompany("Airbnb", "airbnb", "greenhouse")

	s.UpsertPostings(coID, []Posting{
		{ExternalID: "1", Title: "Engineer"},
		{ExternalID: "2", Title: "Designer"},
	})

	all, _ := s.ListPostings(PostingFilter{TitleUnclassified: true})
	s.SetTitleRelevance(all[0].ID, true)
	s.SetTitleRelevance(all[1].ID, false)

	unscored, err := s.ListPostings(PostingFilter{TitleRelevant: BoolPtr(true), Unscored: true})
	if err != nil {
		t.Fatalf("ListPostings title-passed unscored: %v", err)
	}
	if len(unscored) != 1 {
		t.Errorf("expected 1 title-passed unscored posting, got %d", len(unscored))
	}
}
