package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jobhunt/tools/job-monitor/store"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	db = s
	t.Cleanup(func() { db.Close() })
}

func TestImportCSV(t *testing.T) {
	setupTestDB(t)

	csv := "company,ats,interview_count\n" +
		"Acme Corp,Greenhouse,0\n" +
		"Beta Inc,Lever,1\n" +
		"Ghost Co,not-found-2026-10-16,0\n"

	dir := t.TempDir()
	csvPath := filepath.Join(dir, "companies.csv")
	if err := os.WriteFile(csvPath, []byte(csv), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"import", csvPath})

	if err := importCmd.RunE(importCmd, []string{csvPath}); err != nil {
		t.Fatal(err)
	}

	companies, err := db.ListCompanies(store.CompanyFilter{Active: store.BoolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	if len(companies) != 2 {
		t.Fatalf("expected 2 companies, got %d", len(companies))
	}
	if companies[0].Name != "Acme Corp" {
		t.Errorf("expected Acme Corp, got %s", companies[0].Name)
	}
	if companies[0].ATS != "Greenhouse" {
		t.Errorf("expected Greenhouse, got %s", companies[0].ATS)
	}
}

func TestImportDuplicateSkipped(t *testing.T) {
	setupTestDB(t)

	csv := "company,ats,interview_count\nAcme Corp,Greenhouse,0\n"
	dir := t.TempDir()
	csvPath := filepath.Join(dir, "companies.csv")
	os.WriteFile(csvPath, []byte(csv), 0644)

	importCmd.RunE(importCmd, []string{csvPath})

	buf := new(bytes.Buffer)
	importCmd.SetOut(buf)
	importCmd.SetErr(buf)
	if err := importCmd.RunE(importCmd, []string{csvPath}); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "skipped 1 (duplicate)") {
		t.Errorf("expected duplicate skip message, got: %s", output)
	}

	companies, err := db.ListCompanies(store.CompanyFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(companies) != 1 {
		t.Fatalf("expected 1 company after reimport, got %d", len(companies))
	}
}

func TestRemove(t *testing.T) {
	setupTestDB(t)

	db.AddCompany("Test Co", "testco", "Greenhouse")

	buf := new(bytes.Buffer)
	removeCmd.SetOut(buf)
	if err := removeCmd.RunE(removeCmd, []string{"Test Co"}); err != nil {
		t.Fatal(err)
	}

	companies, _ := db.ListCompanies(store.CompanyFilter{Active: store.BoolPtr(true)})
	if len(companies) != 0 {
		t.Fatalf("expected 0 active companies after remove, got %d", len(companies))
	}

	all, _ := db.ListCompanies(store.CompanyFilter{})
	if len(all) != 1 {
		t.Fatal("company should still exist as inactive")
	}
	if all[0].Active {
		t.Error("company should be inactive")
	}
}

func TestListEmpty(t *testing.T) {
	setupTestDB(t)

	buf := new(bytes.Buffer)
	listCmd.SetOut(buf)
	listAll = false
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "No companies found") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}

func TestListActive(t *testing.T) {
	setupTestDB(t)

	db.AddCompany("Alpha", "alpha", "Greenhouse")
	db.AddCompany("Beta", "beta", "Lever")

	buf := new(bytes.Buffer)
	listCmd.SetOut(buf)
	listAll = false
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Alpha") || !strings.Contains(output, "Beta") {
		t.Errorf("expected both companies in output: %s", output)
	}
	if !strings.Contains(output, "Greenhouse") {
		t.Errorf("expected ATS in output: %s", output)
	}
}

func TestListAllShowsInactive(t *testing.T) {
	setupTestDB(t)

	db.AddCompany("Active Co", "activeco", "Greenhouse")
	db.AddCompany("Gone Co", "goneco", "Lever")
	db.DeactivateCompany("Gone Co")

	buf := new(bytes.Buffer)
	listCmd.SetOut(buf)
	listAll = true
	if err := listCmd.RunE(listCmd, nil); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "Active Co") || !strings.Contains(output, "Gone Co") {
		t.Errorf("expected both companies: %s", output)
	}
	if !strings.Contains(output, "ACTIVE") {
		t.Errorf("expected ACTIVE column header: %s", output)
	}
	if !strings.Contains(output, "no") {
		t.Errorf("expected 'no' for inactive company: %s", output)
	}
}
