// Package collector collecte les métriques système en parallèle.
// Il lit les données depuis /proc sur Linux, et utilise runtime Go
// sur les autres systèmes d'exploitation.
package collector

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Metrique représente une mesure système à un instant donné.
type Metrique struct {
	Source string  `json:"source"`
	Valeur float64 `json:"valeur"`
	Unite  string  `json:"unite"`
	Erreur string  `json:"erreur,omitempty"`
}

// Snapshot regroupe toutes les métriques collectées à un instant T.
type Snapshot struct {
	Timestamp time.Time  `json:"timestamp"`
	OS        string     `json:"os"`
	Arch      string     `json:"arch"`
	GoVersion string     `json:"go_version"`
	Metriques []Metrique `json:"metriques"`
}

// source définit une source de métrique avec son propre timeout.
type source struct {
	Nom      string
	Unite    string
	Timeout  time.Duration
	Collecte func(ctx context.Context) (float64, error)
}

// sources retourne la liste des sources disponibles selon l'OS.
func sources() []source {
	srcs := []source{
		{
			Nom:     "CPU Logiques",
			Unite:   "cœurs",
			Timeout: 50 * time.Millisecond,
			Collecte: func(ctx context.Context) (float64, error) {
				return float64(runtime.NumCPU()), nil
			},
		},
		{
			Nom:     "Goroutines",
			Unite:   "actives",
			Timeout: 50 * time.Millisecond,
			Collecte: func(ctx context.Context) (float64, error) {
				return float64(runtime.NumGoroutine()), nil
			},
		},
		{
			Nom:     "Heap Go",
			Unite:   "Mo",
			Timeout: 50 * time.Millisecond,
			Collecte: func(ctx context.Context) (float64, error) {
				var stats runtime.MemStats
				runtime.ReadMemStats(&stats)
				return float64(stats.HeapAlloc) / 1024 / 1024, nil
			},
		},
	}

	// Sources Linux uniquement — lecture depuis /proc
	if runtime.GOOS == "linux" {
		srcs = append(srcs,
			source{
				Nom:     "CPU Usage",
				Unite:   "%",
				Timeout: 1100 * time.Millisecond, // 1 seconde de mesure + marge
				Collecte: func(ctx context.Context) (float64, error) {
					return calculerUsageCPU(ctx, time.Second)
				},
			},
			source{
				Nom:     "RAM Utilisée",
				Unite:   "%",
				Timeout: 100 * time.Millisecond,
				Collecte: func(ctx context.Context) (float64, error) {
					total, dispo, err := lireRAM()
					if err != nil {
						return 0, err
					}
					if total == 0 {
						return 0, fmt.Errorf("RAM totale = 0")
					}
					return float64(total-dispo) / float64(total) * 100, nil
				},
			},
			source{
				Nom:     "RAM Disponible",
				Unite:   "Mo",
				Timeout: 100 * time.Millisecond,
				Collecte: func(ctx context.Context) (float64, error) {
					_, dispo, err := lireRAM()
					return float64(dispo) / 1024 / 1024, err
				},
			},
		)
	}

	return srcs
}

// CollecterSnapshot collecte toutes les métriques en parallèle.
func CollecterSnapshot(ctx context.Context) (Snapshot, error) {
	if ctx.Err() != nil {
		return Snapshot{}, ctx.Err()
	}

	srcs := sources()
	résultats := make(chan Metrique, len(srcs))
	var wg sync.WaitGroup

	for _, src := range srcs {
		wg.Add(1)
		s := src
		go func() {
			defer wg.Done()

			srcCtx, cancel := context.WithTimeout(ctx, s.Timeout)
			defer cancel()

			valeur, err := s.Collecte(srcCtx)

			m := Metrique{
				Source: s.Nom,
				Valeur: valeur,
				Unite:  s.Unite,
			}
			if err != nil {
				m.Erreur = err.Error()
			}
			résultats <- m
		}()
	}

	go func() {
		wg.Wait()
		close(résultats)
	}()

	snap := Snapshot{
		Timestamp: time.Now(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		GoVersion: runtime.Version(),
	}

	for m := range résultats {
		snap.Metriques = append(snap.Metriques, m)
	}

	return snap, nil
}

// --- Helpers Linux /proc ---

// statCPU représente les compteurs CPU de /proc/stat.
type statCPU struct {
	idle  uint64
	total uint64
}

func lireStatCPU() (statCPU, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return statCPU{}, fmt.Errorf("lireStatCPU : %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		ligne := sc.Text()
		if !strings.HasPrefix(ligne, "cpu ") {
			continue
		}

		champs := strings.Fields(ligne)
		if len(champs) < 8 {
			return statCPU{}, fmt.Errorf("/proc/stat format inattendu")
		}

		var vals [7]uint64
		for i := 0; i < 7; i++ {
			v, err := strconv.ParseUint(champs[i+1], 10, 64)
			if err != nil {
				return statCPU{}, fmt.Errorf("parsing /proc/stat : %w", err)
			}
			vals[i] = v
		}

		// user, nice, system, idle, iowait, irq, softirq
		idle  := vals[3] + vals[4]
		total := vals[0] + vals[1] + vals[2] + vals[3] + vals[4] + vals[5] + vals[6]
		return statCPU{idle: idle, total: total}, nil
	}

	return statCPU{}, fmt.Errorf("/proc/stat : ligne cpu non trouvée")
}

func calculerUsageCPU(ctx context.Context, intervalle time.Duration) (float64, error) {
	avant, err := lireStatCPU()
	if err != nil {
		return 0, err
	}

	select {
	case <-time.After(intervalle):
	case <-ctx.Done():
		return 0, ctx.Err()
	}

	après, err := lireStatCPU()
	if err != nil {
		return 0, err
	}

	idleDelta := float64(après.idle - avant.idle)
	totalDelta := float64(après.total - avant.total)

	if totalDelta == 0 {
		return 0, nil
	}

	return (1 - idleDelta/totalDelta) * 100, nil
}

func lireRAM() (total, disponible uint64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, fmt.Errorf("lireRAM : %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		champs := strings.Fields(sc.Text())
		if len(champs) < 2 {
			continue
		}
		val, err := strconv.ParseUint(champs[1], 10, 64)
		if err != nil {
			continue
		}
		switch champs[0] {
		case "MemTotal:":
			total = val * 1024
		case "MemAvailable:":
			disponible = val * 1024
		}
	}

	return total, disponible, sc.Err()
}
