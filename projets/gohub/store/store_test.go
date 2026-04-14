package store

import (
	"sync"
	"testing"
	"time"
)

func TestStore_VideInitialement(t *testing.T) {
	s := New()
	_, ok := s.Dernier()
	if ok {
		t.Error("un store vide ne devrait pas avoir de dernier élément")
	}
	if n := s.Compter(); n != 0 {
		t.Errorf("Compter() attendu 0, obtenu %d", n)
	}
}

func TestStore_AjouterAssigneID(t *testing.T) {
	s := New()

	for i := 1; i <= 3; i++ {
		s.Ajouter(Snapshot{Timestamp: time.Now()})
		dernier, ok := s.Dernier()
		if !ok {
			t.Fatalf("itération %d : Dernier() devrait retourner true", i)
		}
		if int(dernier.ID) != i {
			t.Errorf("itération %d : ID attendu %d, obtenu %d", i, i, dernier.ID)
		}
	}
}

func TestStore_Historique(t *testing.T) {
	tests := []struct {
		nom     string
		inserts int
		limite  int
		attendu int
	}{
		{"limite < total",   10, 5,   5},
		{"limite > total",   3,  10,  3},
		{"store vide",       0,  5,   0},
		{"limite zéro",      5,  0,   0},
		{"limite négative",  5,  -1,  0},
		{"limite = total",   5,  5,   5},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.nom, func(t *testing.T) {
			s := New()
			for i := 0; i < tc.inserts; i++ {
				s.Ajouter(Snapshot{Timestamp: time.Now()})
			}
			if got := len(s.Historique(tc.limite)); got != tc.attendu {
				t.Errorf("Historique(%d) avec %d inserts : attendu %d, obtenu %d",
					tc.limite, tc.inserts, tc.attendu, got)
			}
		})
	}
}

func TestStore_HistoriqueOrdreChronologique(t *testing.T) {
	s := New()
	for i := 0; i < 5; i++ {
		s.Ajouter(Snapshot{Timestamp: time.Now()})
	}

	hist := s.Historique(5)
	for i := 1; i < len(hist); i++ {
		if hist[i].ID <= hist[i-1].ID {
			t.Errorf("ordre incorrect : hist[%d].ID=%d <= hist[%d].ID=%d",
				i, hist[i].ID, i-1, hist[i-1].ID)
		}
	}
}

func TestStore_LimiteMaximale(t *testing.T) {
	s := New()
	for i := 0; i < maxSnapshots+100; i++ {
		s.Ajouter(Snapshot{Timestamp: time.Now()})
	}
	if n := s.Compter(); n > maxSnapshots {
		t.Errorf("store dépasse la limite : %d > %d", n, maxSnapshots)
	}
}

func TestStore_ConcurrenceSafe(t *testing.T) {
	t.Parallel()

	s := New()
	var wg sync.WaitGroup
	const n = 500

	// Écritures concurrentes
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Ajouter(Snapshot{Timestamp: time.Now()})
		}()
	}

	// Lectures concurrentes pendant les écritures
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Dernier()
			s.Historique(10)
			s.Compter()
		}()
	}

	wg.Wait()

	if got := s.Compter(); got != n {
		t.Errorf("attendu %d snapshots, obtenu %d", n, got)
	}
}

func BenchmarkStore_Ajouter(b *testing.B) {
	s := New()
	snap := Snapshot{
		Timestamp: time.Now(),
		Metriques: []Metrique{
			{Source: "CPU", Valeur: 23.4, Unite: "%"},
			{Source: "RAM", Valeur: 67.1, Unite: "%"},
		},
	}
	for i := 0; i < b.N; i++ {
		s.Ajouter(snap)
	}
}

func BenchmarkStore_Dernier(b *testing.B) {
	s := New()
	for i := 0; i < 100; i++ {
		s.Ajouter(Snapshot{Timestamp: time.Now()})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Dernier()
	}
}

func BenchmarkStore_Historique100(b *testing.B) {
	s := New()
	for i := 0; i < 1000; i++ {
		s.Ajouter(Snapshot{Timestamp: time.Now()})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Historique(100)
	}
}
