// Package store fournit un stockage en mémoire thread-safe pour les snapshots.
// Il utilise sync.RWMutex pour permettre des lectures simultanées
// tout en garantissant l'exclusivité des écritures.
package store

import (
	"sync"
	"time"
)

const maxSnapshots = 1000

// Metrique représente une mesure système à un instant donné.
type Metrique struct {
	Source string  `json:"source"`
	Valeur float64 `json:"valeur"`
	Unite  string  `json:"unite"`
	Erreur string  `json:"erreur,omitempty"`
}

// Snapshot regroupe les métriques collectées à un instant T.
type Snapshot struct {
	ID        int64      `json:"id"`
	Timestamp time.Time  `json:"timestamp"`
	OS        string     `json:"os,omitempty"`
	Arch      string     `json:"arch,omitempty"`
	GoVersion string     `json:"go_version,omitempty"`
	Metriques []Metrique `json:"metriques"`
}

// Store est un dépôt de snapshots en mémoire, thread-safe.
type Store struct {
	mu        sync.RWMutex
	snapshots []Snapshot
	nextID    int64
}

// New crée et retourne un Store vide.
func New() *Store {
	return &Store{}
}

// Ajouter insère un snapshot dans le store.
// L'ID est assigné automatiquement.
// Le store est limité à maxSnapshots entrées — les plus anciennes sont supprimées.
func (s *Store) Ajouter(snap Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	snap.ID = s.nextID

	s.snapshots = append(s.snapshots, snap)

	// Écrêtage — on garde uniquement les N derniers snapshots
	if len(s.snapshots) > maxSnapshots {
		s.snapshots = s.snapshots[len(s.snapshots)-maxSnapshots:]
	}
}

// Dernier retourne le snapshot le plus récent.
// Retourne false si le store est vide.
func (s *Store) Dernier() (Snapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.snapshots) == 0 {
		return Snapshot{}, false
	}
	return s.snapshots[len(s.snapshots)-1], true
}

// Historique retourne les N derniers snapshots dans l'ordre chronologique.
// Si limite > nombre de snapshots disponibles, retourne tous les snapshots.
func (s *Store) Historique(limite int) []Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.snapshots)
	if limite <= 0 || total == 0 {
		return nil
	}
	if limite > total {
		limite = total
	}

	résultat := make([]Snapshot, limite)
	copy(résultat, s.snapshots[total-limite:])
	return résultat
}

// Compter retourne le nombre de snapshots stockés.
func (s *Store) Compter() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.snapshots)
}
