package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/votre-pseudo/gowatch/collector"
	"github.com/votre-pseudo/gowatch/exporter"
	"github.com/votre-pseudo/gowatch/renderer"
	"github.com/votre-pseudo/gowatch/scanner"
)

// Variables injectées à la compilation via -ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func main() {
	// --- Flags CLI ---
	showVersion := flag.Bool("version", false, "Afficher la version")
	format      := flag.String("format", "text", "Format de sortie : text | json")
	watch       := flag.Bool("watch", false, "Mode surveillance — rafraîchissement continu")
	interval    := flag.Duration("interval", 2*time.Second, "Intervalle de rafraîchissement (mode --watch)")
	scanHote    := flag.String("scan", "", "Hôte à scanner (ex: localhost)")
	scanPorts   := flag.String("ports", "1-1024", "Plage de ports à scanner (ex: 1-1024)")
	exportURL   := flag.String("export", "", "URL HTTPS d'export des métriques vers gohub")
	skipVerify  := flag.Bool("skip-verify", false, "Ignorer la vérification TLS (tests uniquement)")
	flag.Parse()

	// --- Version ---
	if *showVersion {
		fmt.Printf("gowatch %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
		os.Exit(0)
	}

	// --- Context racine — annulé sur Ctrl+C ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		fmt.Fprintf(os.Stderr, "\nSignal reçu (%s) — arrêt propre...\n", sig)
		cancel()
	}()

	// --- Choisir le renderer ---
	var rnd renderer.Renderer
	switch *format {
	case "json":
		rnd = renderer.NewJSON(os.Stdout)
	default:
		rnd = renderer.NewText(os.Stdout)
	}

	// --- Scanner de ports (si demandé) ---
	if *scanHote != "" {
		var portDebut, portFin int
		fmt.Sscanf(*scanPorts, "%d-%d", &portDebut, &portFin)

		fmt.Fprintf(os.Stderr, "Scan de %s (ports %d-%d)...\n", *scanHote, portDebut, portFin)
		scanCtx, scanCancel := context.WithTimeout(ctx, 60*time.Second)
		defer scanCancel()

		debut := time.Now()
		résultats := scanner.Scan(scanCtx, *scanHote, portDebut, portFin, 200)
		duree := time.Since(debut)

		ouverts := 0
		for _, r := range résultats {
			if r.Ouvert {
				ouverts++
				fmt.Printf("  ✓ Port %-6d : %s\n", r.Port, scanner.NomService(r.Port))
			}
		}
		fmt.Printf("\n%d ports ouverts — scan terminé en %v\n\n",
			ouverts, duree.Round(time.Millisecond))
	}

	// --- Mode watch ou snapshot unique ---
	snapshots := 0

	collecterEtAfficher := func() bool {
		snap, err := collector.CollecterSnapshot(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return false // Contexte annulé — arrêt normal
			}
			fmt.Fprintf(os.Stderr, "Erreur de collecte : %v\n", err)
			return true
		}

		snapshots++
		if err := rnd.Render(snap); err != nil {
			fmt.Fprintf(os.Stderr, "Erreur de rendu : %v\n", err)
		}

		// Export si demandé
		if *exportURL != "" {
			if err := exporter.ExporterHTTPS(*exportURL, snap, *skipVerify); err != nil {
				fmt.Fprintf(os.Stderr, "Erreur d'export : %v\n", err)
			}
		}

		return true
	}

	if *watch {
		fmt.Fprintf(os.Stderr, "Mode surveillance (intervalle: %v) — Ctrl+C pour arrêter\n\n", *interval)
		ticker := time.NewTicker(*interval)
		defer ticker.Stop()

		collecterEtAfficher() // Premier snapshot immédiat

		for {
			select {
			case <-ticker.C:
				if !collecterEtAfficher() {
					goto fin
				}
			case <-ctx.Done():
				goto fin
			}
		}
	} else {
		collecterEtAfficher()
	}

fin:
	if *watch {
		fmt.Fprintf(os.Stderr, "Arrêt propre. %d snapshot(s) collecté(s).\n", snapshots)
	}
}
