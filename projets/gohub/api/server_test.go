package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/votre-pseudo/gohub/store"
)

// setupServeur crée un Server de test avec des données initiales.
func setupServeur(t *testing.T) *Server {
	t.Helper()

	s := store.New()
	s.Ajouter(store.Snapshot{
		Timestamp: time.Now(),
		Metriques: []store.Metrique{
			{Source: "CPU", Valeur: 23.4, Unite: "%"},
			{Source: "RAM", Valeur: 67.1, Unite: "%"},
		},
	})

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Silencieux pendant les tests
	}))

	return NewServer(s, "test-1.0.0", logger)
}

func TestHandleHealth(t *testing.T) {
	srv := setupServeur(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.handleHealth(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status attendu %d, obtenu %d", http.StatusOK, rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("décodage JSON impossible : %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("status attendu 'ok', obtenu '%s'", body["status"])
	}
	if body["version"] != "test-1.0.0" {
		t.Errorf("version attendue 'test-1.0.0', obtenu '%s'", body["version"])
	}
	if body["uptime"] == "" {
		t.Error("uptime ne devrait pas être vide")
	}
}

func TestHandleMetrics_Méthodes(t *testing.T) {
	tests := []struct {
		méthode       string
		statusAttendu int
	}{
		{http.MethodGet,    http.StatusOK},
		{http.MethodPost,   http.StatusMethodNotAllowed},
		{http.MethodDelete, http.StatusMethodNotAllowed},
		{http.MethodPut,    http.StatusMethodNotAllowed},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.méthode, func(t *testing.T) {
			srv := setupServeur(t)
			req := httptest.NewRequest(tc.méthode, "/api/metrics", nil)
			rec := httptest.NewRecorder()

			srv.handleMetrics(rec, req)

			if rec.Code != tc.statusAttendu {
				t.Errorf("%s : status attendu %d, obtenu %d",
					tc.méthode, tc.statusAttendu, rec.Code)
			}
		})
	}
}

func TestHandleMetrics_StoreVide(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	srv := NewServer(store.New(), "test", logger)

	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	rec := httptest.NewRecorder()
	srv.handleMetrics(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("store vide : status attendu %d, obtenu %d",
			http.StatusNotFound, rec.Code)
	}
}

func TestHandleHistory_Limite(t *testing.T) {
	tests := []struct {
		nom           string
		queryLimite   string
		statusAttendu int
		countAttendu  int
	}{
		{"défaut (10)",     "",    http.StatusOK, 1},
		{"limite=1",        "1",   http.StatusOK, 1},
		{"limite invalide", "abc", http.StatusBadRequest, 0},
		{"limite négative", "-1",  http.StatusBadRequest, 0},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.nom, func(t *testing.T) {
			srv := setupServeur(t)
			url := "/api/metrics/history"
			if tc.queryLimite != "" {
				url += "?limit=" + tc.queryLimite
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			rec := httptest.NewRecorder()
			srv.handleHistory(rec, req)

			if rec.Code != tc.statusAttendu {
				t.Errorf("%s : status attendu %d, obtenu %d",
					tc.nom, tc.statusAttendu, rec.Code)
			}
		})
	}
}

func TestHandleIngest(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	s := store.New()
	srv := NewServer(s, "test", logger)

	snap := store.Snapshot{
		Metriques: []store.Metrique{
			{Source: "CPU", Valeur: 42.0, Unite: "%"},
		},
	}
	corps, _ := json.Marshal(snap)

	req := httptest.NewRequest(http.MethodPost, "/api/metrics/ingest",
		bytes.NewReader(corps))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.handleIngest(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status attendu %d, obtenu %d", http.StatusCreated, rec.Code)
	}

	if s.Compter() != 1 {
		t.Errorf("store devrait contenir 1 snapshot, contient %d", s.Compter())
	}
}

func TestHandleIngest_JSONInvalide(t *testing.T) {
	srv := setupServeur(t)
	req := httptest.NewRequest(http.MethodPost, "/api/metrics/ingest",
		bytes.NewReader([]byte("pas du json")))
	rec := httptest.NewRecorder()

	srv.handleIngest(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("JSON invalide : status attendu %d, obtenu %d",
			http.StatusBadRequest, rec.Code)
	}
}

func TestContentTypeJSON(t *testing.T) {
	srv := setupServeur(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	srv.handleHealth(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct == "" {
		t.Error("Content-Type ne devrait pas être vide")
	}
}
