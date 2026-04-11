package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof" // Expose les endpoints pprof sur le port debug
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/votre-pseudo/gohub/api"
	"github.com/votre-pseudo/gohub/store"
)

// Variables injectées à la compilation via -ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	// --- Flags ---
	port       := flag.String("port", "8080", "Port d'écoute HTTP")
	debugPort  := flag.String("debug-port", "6060", "Port pprof (localhost uniquement)")
	logFormat  := flag.String("log-format", "text", "Format des logs : text | json")
	showVer    := flag.Bool("version", false, "Afficher la version et quitter")
	flag.Parse()

	if *showVer {
		fmt.Printf("gohub %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		os.Exit(0)
	}

	// --- Logger structuré ---
	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	if *logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	logger := slog.New(handler)

	logger.Info("démarrage gohub",
		slog.String("version",    Version),
		slog.String("commit",     Commit),
		slog.String("build_date", BuildDate),
		slog.String("go_version", runtime.Version()),
		slog.String("os",         runtime.GOOS+"/"+runtime.GOARCH),
		slog.Int("cpu_count",     runtime.NumCPU()),
	)

	// --- Store ---
	s := store.New()

	// Snapshot initial de bienvenue
	s.Ajouter(store.Snapshot{
		Timestamp: time.Now(),
		Metriques: []store.Metrique{
			{Source: "CPU",        Valeur: float64(runtime.NumCPU()),       Unite: "cœurs"},
			{Source: "Goroutines", Valeur: float64(runtime.NumGoroutine()), Unite: "actives"},
		},
	})

	// --- Serveur pprof (localhost uniquement) ---
	go func() {
		addr := "localhost:" + *debugPort
		logger.Info("pprof disponible", slog.String("addr", addr))
		if err := http.ListenAndServe(addr, nil); err != nil {
			logger.Warn("pprof arrêté", slog.String("erreur", err.Error()))
		}
	}()

	// --- Serveur API ---
	srv := api.NewServer(s, Version, logger)

	httpSrv := &http.Server{
		Addr:         ":" + *port,
		Handler:      srv,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Démarrer en arrière-plan
	go func() {
		logger.Info("serveur HTTP démarré", slog.String("addr", ":"+*port))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("erreur serveur", slog.String("erreur", err.Error()))
			os.Exit(1)
		}
	}()

	// --- Graceful shutdown ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	logger.Info("signal reçu — arrêt gracieux",
		slog.String("signal", sig.String()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Error("erreur shutdown", slog.String("erreur", err.Error()))
		os.Exit(1)
	}

	logger.Info("serveur arrêté proprement")
}
