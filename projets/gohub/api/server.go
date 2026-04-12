// Package api implémente le serveur HTTP de gohub.
// Il expose les endpoints REST pour consulter et ingérer des métriques.
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/votre-pseudo/gohub/store"
)

// Server est le serveur HTTP de gohub.
// Il implémente http.Handler via ServeHTTP.
type Server struct {
	store     *store.Store
	version   string
	logger    *slog.Logger
	startTime time.Time
	mux       *http.ServeMux
}

// NewServer crée un Server et enregistre toutes les routes.
func NewServer(s *store.Store, version string, logger *slog.Logger) *Server {
	srv := &Server{
		store:     s,
		version:   version,
		logger:    logger,
		startTime: time.Now(),
		mux:       http.NewServeMux(),
	}
	srv.enregistrerRoutes()
	return srv
}

// ServeHTTP implémente http.Handler.
// Il applique le middleware de logging avant de déléguer au routeur.
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	debut := time.Now()

	// Wrapper pour capturer le status code
	rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}

	// Middleware recover — empêche un panic de tuer le serveur
	defer func() {
		if err := recover(); err != nil {
			srv.logger.Error("panic récupéré",
				slog.Any("erreur", err),
				slog.String("path", r.URL.Path),
			)
			srv.erreur(w, http.StatusInternalServerError, "Erreur interne du serveur")
		}
	}()

	srv.mux.ServeHTTP(rw, r)

	srv.logger.Info("http",
		slog.String("method",  r.Method),
		slog.String("path",    r.URL.Path),
		slog.Int("status",     rw.code),
		slog.Duration("dur",   time.Since(debut)),
		slog.String("ip",      r.RemoteAddr),
	)
}

// enregistrerRoutes déclare tous les endpoints de l'API.
func (srv *Server) enregistrerRoutes() {
	srv.mux.HandleFunc("/health",                srv.handleHealth)
	srv.mux.HandleFunc("/api/metrics",           srv.handleMetrics)
	srv.mux.HandleFunc("/api/metrics/history",   srv.handleHistory)
	srv.mux.HandleFunc("/api/metrics/ingest",    srv.handleIngest)
}

// --- Helpers ---

// json envoie une réponse JSON avec le code HTTP donné.
func (srv *Server) json(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		srv.logger.Warn("erreur encodage JSON", slog.String("erreur", err.Error()))
	}
}

// erreur envoie une réponse d'erreur JSON standardisée.
func (srv *Server) erreur(w http.ResponseWriter, status int, message string) {
	srv.json(w, status, map[string]interface{}{
		"code":    status,
		"message": message,
	})
}

// méthodeRequise vérifie la méthode HTTP et retourne false si elle ne correspond pas.
func (srv *Server) méthodeRequise(w http.ResponseWriter, r *http.Request, méthode string) bool {
	if r.Method != méthode {
		srv.erreur(w, http.StatusMethodNotAllowed,
			"Méthode "+r.Method+" non autorisée — attendu "+méthode)
		return false
	}
	return true
}

// --- Handlers ---

// handleHealth — GET /health
// Retourne le statut du service, la version, et l'uptime.
func (srv *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	srv.json(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": srv.version,
		"uptime":  time.Since(srv.startTime).Round(time.Second).String(),
	})
}

// handleMetrics — GET /api/metrics
// Retourne le dernier snapshot collecté.
func (srv *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if !srv.méthodeRequise(w, r, http.MethodGet) {
		return
	}

	snap, ok := srv.store.Dernier()
	if !ok {
		srv.erreur(w, http.StatusNotFound, "Aucune métrique disponible — en attente du premier snapshot")
		return
	}

	srv.json(w, http.StatusOK, snap)
}

// handleHistory — GET /api/metrics/history?limit=N
// Retourne les N derniers snapshots (max 100, défaut 10).
func (srv *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if !srv.méthodeRequise(w, r, http.MethodGet) {
		return
	}

	limite := 10
	if l := r.URL.Query().Get("limit"); l != "" {
		n, err := strconv.Atoi(l)
		if err != nil || n <= 0 {
			srv.erreur(w, http.StatusBadRequest, "Paramètre 'limit' invalide — doit être un entier positif")
			return
		}
		if n > 100 {
			n = 100 // Plafond de sécurité
		}
		limite = n
	}

	snapshots := srv.store.Historique(limite)
	srv.json(w, http.StatusOK, map[string]interface{}{
		"count":      len(snapshots),
		"total":      srv.store.Compter(),
		"limite":     limite,
		"snapshots":  snapshots,
	})
}

// handleIngest — POST /api/metrics/ingest
// Accepte un snapshot JSON et le stocke.
// Utilisé par gowatch pour envoyer ses métriques.
func (srv *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if !srv.méthodeRequise(w, r, http.MethodPost) {
		return
	}

	// Limiter la taille du body — protection contre les requêtes géantes
	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576) // 1 Mo max

	var snap store.Snapshot
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&snap); err != nil {
		srv.erreur(w, http.StatusBadRequest, "JSON invalide : "+err.Error())
		return
	}

	// Écraser le timestamp avec l'heure de réception — source de vérité côté serveur
	snap.Timestamp = time.Now()

	srv.store.Ajouter(snap)

	srv.logger.Info("snapshot ingéré",
		slog.Int("metriques", len(snap.Metriques)),
		slog.Int64("id", snap.ID),
	)

	srv.json(w, http.StatusCreated, map[string]interface{}{
		"message": "Snapshot enregistré",
		"total":   srv.store.Compter(),
	})
}

// --- responseWriter wrapper ---

// responseWriter capture le status code pour le logging.
type responseWriter struct {
	http.ResponseWriter
	code int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.code = code
	rw.ResponseWriter.WriteHeader(code)
}
