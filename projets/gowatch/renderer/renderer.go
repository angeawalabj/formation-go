// Package renderer fournit les implémentations d'affichage des snapshots.
// L'interface Renderer permet d'ajouter facilement de nouveaux formats.
package renderer

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/angeawalabj/gowatch/collector"
)

// Renderer définit comment afficher un Snapshot.
// Tout type implémentant Render() satisfait cette interface.
type Renderer interface {
	Render(snap collector.Snapshot) error
}

// --- JSON Renderer ---

// JSONRenderer affiche le snapshot en JSON indenté.
type JSONRenderer struct {
	w io.Writer
}

// NewJSON crée un JSONRenderer qui écrit dans w.
func NewJSON(w io.Writer) *JSONRenderer {
	return &JSONRenderer{w: w}
}

func (r *JSONRenderer) Render(snap collector.Snapshot) error {
	enc := json.NewEncoder(r.w)
	enc.SetIndent("", "  ")
	return enc.Encode(snap)
}

// --- Text Renderer ---

// TextRenderer affiche le snapshot en format texte lisible.
type TextRenderer struct {
	w     io.Writer
	first bool
}

// NewText crée un TextRenderer qui écrit dans w.
func NewText(w io.Writer) *TextRenderer {
	return &TextRenderer{w: w, first: true}
}

func (r *TextRenderer) Render(snap collector.Snapshot) error {
	// Séparateur entre snapshots en mode watch
	if !r.first {
		fmt.Fprintln(r.w, strings.Repeat("─", 42))
	}
	r.first = false

	// En-tête
	fmt.Fprintf(r.w, "=== gowatch · %s ===\n", snap.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(r.w, "  OS      : %s/%s\n", snap.OS, snap.Arch)
	fmt.Fprintf(r.w, "  Go      : %s\n\n", snap.GoVersion)

	// Métriques
	fmt.Fprintln(r.w, "Métriques :")
	for _, m := range snap.Metriques {
		if m.Erreur != "" {
			fmt.Fprintf(r.w, "  ✗ %-18s : erreur (%s)\n", m.Source, m.Erreur)
		} else {
			fmt.Fprintf(r.w, "  ✓ %-18s : %.2f %s\n", m.Source, m.Valeur, m.Unite)
		}
	}
	fmt.Fprintln(r.w)

	return nil
}
