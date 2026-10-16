package atsutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCheckProviderOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	p := Provider{Name: "Test", BaseURL: srv.URL + "/", MinSize: 0}
	if !CheckProvider(srv.Client(), p, "testslug") {
		t.Error("expected CheckProvider to return true for 200 response")
	}
}

func TestCheckProvider404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()

	p := Provider{Name: "Test", BaseURL: srv.URL + "/", MinSize: 0}
	if CheckProvider(srv.Client(), p, "badslug") {
		t.Error("expected CheckProvider to return false for 404 response")
	}
}

func TestCheckProviderMinSizePass(t *testing.T) {
	body := strings.Repeat("x", 16000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	p := Provider{Name: "Ashby", BaseURL: srv.URL + "/", MinSize: 15000}
	if !CheckProvider(srv.Client(), p, "slug") {
		t.Error("expected CheckProvider to return true when body exceeds MinSize")
	}
}

func TestCheckProviderMinSizeFail(t *testing.T) {
	body := strings.Repeat("x", 100)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(body))
	}))
	defer srv.Close()

	p := Provider{Name: "Ashby", BaseURL: srv.URL + "/", MinSize: 15000}
	if CheckProvider(srv.Client(), p, "slug") {
		t.Error("expected CheckProvider to return false when body is below MinSize")
	}
}

func TestCheckProviderTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	client := &http.Client{Timeout: 100 * time.Millisecond}
	p := Provider{Name: "Slow", BaseURL: srv.URL + "/", MinSize: 0}
	if CheckProvider(client, p, "slug") {
		t.Error("expected CheckProvider to return false on timeout")
	}
}

func TestDetectATS(t *testing.T) {
	// Set up fake servers for each provider
	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/airbnb" {
			w.WriteHeader(200)
			return
		}
		w.WriteHeader(404)
	}))
	defer ghSrv.Close()

	leverSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer leverSrv.Close()

	ashbySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ashbySrv.Close()

	// Override Providers for the test
	origProviders := Providers
	Providers = []Provider{
		{Name: "Greenhouse", BaseURL: ghSrv.URL + "/", MinSize: 0},
		{Name: "Lever", BaseURL: leverSrv.URL + "/", MinSize: 0},
		{Name: "Ashby", BaseURL: ashbySrv.URL + "/", MinSize: 15000},
	}
	defer func() { Providers = origProviders }()

	result := DetectATS(http.DefaultClient, "Airbnb")
	if result != "Greenhouse" {
		t.Errorf("expected 'Greenhouse', got %q", result)
	}
}

func TestDetectATSNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer srv.Close()

	origProviders := Providers
	Providers = []Provider{
		{Name: "Greenhouse", BaseURL: srv.URL + "/", MinSize: 0},
		{Name: "Lever", BaseURL: srv.URL + "/", MinSize: 0},
		{Name: "Ashby", BaseURL: srv.URL + "/", MinSize: 15000},
	}
	defer func() { Providers = origProviders }()

	result := DetectATS(http.DefaultClient, "UnknownCorp")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
