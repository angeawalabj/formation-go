# Chapitre 5.2 — Optimisation Bas Niveau

> *"Premature optimization is the root of all evil."*
> — Donald Knuth
>
> *"Mais une optimisation bien mesurée, au bon endroit,
> c'est de l'ingénierie."*
> — La communauté Go

---

## Le problème

Votre programme Go est lent. Ou il consomme trop de mémoire. Ou les deux.

La réaction instinctive est de réécrire les parties qui "semblent" lentes. C'est presque toujours une erreur. Dans la majorité des cas, le goulot d'étranglement réel est ailleurs — dans une allocation mémoire répétée inutilement, dans un verrou trop large, dans une désérialisation JSON coûteuse dans une boucle chaude.

Go vous donne les outils pour **mesurer d'abord, optimiser ensuite**. Ce chapitre vous montre comment les utiliser — et quand descendre vraiment bas niveau avec `unsafe` et les syscalls.

---

## L'intuition

### Le cycle d'optimisation Go

```
1. Mesurer          → pprof, benchmarks
2. Identifier       → où est vraiment le goulot ?
3. Comprendre       → pourquoi c'est lent ?
4. Optimiser        → changer le code
5. Remesurer        → est-ce que ça a vraiment amélioré ?
6. Recommencer      → jusqu'au niveau acceptable
```

Ne jamais sauter l'étape 1. Ne jamais sauter l'étape 5.

### Les deux sources de lenteur en Go

**CPU** — votre code fait trop de calculs. Le profiler CPU vous montre quelles fonctions consomment le plus de temps processeur.

**Mémoire (allocations)** — votre code crée trop d'objets. Le garbage collector passe son temps à les nettoyer. Moins d'allocations = moins de GC = plus de vitesse.

En pratique, la majorité des problèmes de performance Go viennent des **allocations mémoire excessives**, pas de calculs trop lourds.

---

## La solution Go

### Profiling CPU avec `pprof`

On a vu au Chapitre 4.3 comment exposer `pprof` via HTTP. On peut aussi profiler directement dans le code — utile pour les programmes CLI ou les benchmarks.

```go
package main

import (
    "os"
    "runtime/pprof"
    "time"
    "log"
)

func travailIntensif() {
    // Simuler un calcul coûteux
    résultat := 0
    for i := 0; i < 10_000_000; i++ {
        résultat += i * i
    }
    _ = résultat
}

func main() {
    // Démarrer le profiling CPU
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal(err)
    }
    defer pprof.StopCPUProfile()

    // Code à profiler
    for i := 0; i < 10; i++ {
        travailIntensif()
    }

    // Profil mémoire — snapshot à un instant donné
    mf, _ := os.Create("mem.prof")
    defer mf.Close()
    pprof.WriteHeapProfile(mf)
}
```

**Analyser les profils :**
```bash
# Lancer le programme — génère cpu.prof et mem.prof
go run main.go

# Analyser le profil CPU — interface web interactive
go tool pprof -http=:8090 cpu.prof
# Ouvrez http://localhost:8090

# Analyser en ligne de commande
go tool pprof cpu.prof
# (pprof) top10          ← Les 10 fonctions les plus coûteuses
# (pprof) list maFonction ← Code source annoté avec les temps
# (pprof) web            ← Graphe de flamme dans le navigateur

# Analyser la mémoire
go tool pprof mem.prof
# (pprof) top10 -cum     ← Allocations cumulatives
# (pprof) list maFonction
```

**Lire un flamegraph `pprof` :**
- Chaque rectangle = une fonction
- La largeur = le pourcentage de temps CPU consommé
- La hauteur = la profondeur de la pile d'appels
- Les rectangles larges en bas de pile = les goulots d'étranglement réels

---

### Profiling dans les benchmarks

La façon la plus précise de profiler une fonction spécifique :

```go
// Benchmarks dans un fichier _test.go
func BenchmarkTravailIntensif(b *testing.B) {
    for i := 0; i < b.N; i++ {
        travailIntensif()
    }
}
```

```bash
# Benchmark avec profil CPU
go test -bench=BenchmarkTravailIntensif -cpuprofile=cpu.prof ./...
go tool pprof -http=:8090 cpu.prof

# Benchmark avec profil mémoire
go test -bench=BenchmarkTravailIntensif -memprofile=mem.prof ./...
go tool pprof -http=:8090 mem.prof
```

---

### Réduire les allocations — La principale optimisation Go

Chaque fois que Go crée un objet sur le tas (heap), il devra éventuellement le nettoyer via le garbage collector. Moins d'allocations = moins de GC = programmes plus rapides et plus prévisibles.

**Technique 1 — `sync.Pool` : réutiliser des objets coûteux**

```go
import "sync"

// Pool d'objets réutilisables — évite les allocations répétées
var bufferPool = sync.Pool{
    New: func() interface{} {
        // Créé seulement si le pool est vide
        return make([]byte, 0, 4096)
    },
}

func traiterRequete(données []byte) []byte {
    // Obtenir un buffer depuis le pool
    buf := bufferPool.Get().([]byte)
    buf = buf[:0] // Réinitialiser sans réallouer

    defer bufferPool.Put(buf) // Remettre dans le pool après usage

    // Utiliser le buffer...
    buf = append(buf, données...)
    buf = append(buf, " traité"...)

    résultat := make([]byte, len(buf))
    copy(résultat, buf)
    return résultat
}
```

> 💡 **`sync.Pool` et le GC** — Le pool peut être vidé par le garbage collector à tout moment. Ne stockez pas d'état important dedans — uniquement des objets réutilisables et sans état (buffers, slices, builders).

---

**Technique 2 — Pré-allouer les slices**

```go
// ❌ Croissance par append — réallocations multiples
func sansPréallocation(n int) []int {
    var résultat []int
    for i := 0; i < n; i++ {
        résultat = append(résultat, i*i) // Peut réallouer plusieurs fois
    }
    return résultat
}

// ✅ Pré-allocation — une seule allocation
func avecPréallocation(n int) []int {
    résultat := make([]int, 0, n) // Capacité = n dès le départ
    for i := 0; i < n; i++ {
        résultat = append(résultat, i*i) // Jamais de réallocation
    }
    return résultat
}
```

**Benchmark comparatif :**
```go
func BenchmarkSansPréallocation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        sansPréallocation(10000)
    }
}

func BenchmarkAvecPréallocation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        avecPréallocation(10000)
    }
}
```

```bash
go test -bench=. -benchmem ./...
# BenchmarkSansPréallocation-8    8234    145823 ns/op    357632 B/op    18 allocs/op
# BenchmarkAvecPréallocation-8   18291     65401 ns/op     81920 B/op     1 allocs/op
# ↑ 2.2x plus rapide, 4.4x moins d'allocations — juste avec make([]int, 0, n)
```

---

**Technique 3 — `strings.Builder` pour la concaténation de strings**

```go
import "strings"

// ❌ Concaténation naïve — crée une nouvelle string à chaque +
func concaténerNaïf(mots []string) string {
    résultat := ""
    for _, mot := range mots {
        résultat += mot + " " // Nouvelle allocation à chaque itération !
    }
    return résultat
}

// ✅ strings.Builder — une seule allocation finale
func concaténerEfficace(mots []string) string {
    var sb strings.Builder
    sb.Grow(len(mots) * 10) // Estimation de la taille finale — optionnel mais utile
    for _, mot := range mots {
        sb.WriteString(mot)
        sb.WriteByte(' ')
    }
    return sb.String()
}
```

---

**Technique 4 — Passer par pointeur les grandes structs**

```go
type GrandeStruct struct {
    Champ1 [1000]byte
    Champ2 [1000]byte
    // ...
}

// ❌ Passe une copie de 2000 bytes à chaque appel
func traiterParValeur(s GrandeStruct) {
    // ...
}

// ✅ Passe seulement un pointeur (8 bytes sur 64 bits)
func traiterParPointeur(s *GrandeStruct) {
    // ...
}
```

---

### `runtime` — Inspecter le moteur interne

```go
import "runtime"

func afficherStatsRuntime() {
    var stats runtime.MemStats
    runtime.ReadMemStats(&stats)

    fmt.Printf("Heap allouée       : %d Mo\n", stats.HeapAlloc/1024/1024)
    fmt.Printf("Heap totale sys    : %d Mo\n", stats.HeapSys/1024/1024)
    fmt.Printf("Cycles GC          : %d\n",   stats.NumGC)
    fmt.Printf("Pause GC totale    : %v\n",   time.Duration(stats.PauseTotalNs))
    fmt.Printf("Goroutines actives : %d\n",   runtime.NumGoroutine())
}

// Forcer un cycle GC — utile dans les benchmarks pour partir d'un état propre
runtime.GC()

// Lire les stats GOGC — contrôle la fréquence du GC
// GOGC=100 (défaut) : GC quand la heap double
// GOGC=200 : GC moins fréquent, plus de mémoire utilisée
// GOGC=off : désactiver le GC (dangereux !)
```

---

### `unsafe` — Sortir des rails de Go

`unsafe` est le package qui permet de contourner le système de types de Go. Il donne accès à des opérations que le compilateur refuse normalement : arithmétique de pointeurs, conversion entre types incompatibles, lecture directe de la mémoire.

```go
import (
    "fmt"
    "unsafe"
)

// Taille d'un type en mémoire
type MaStruct struct {
    A int32   // 4 bytes
    B bool    // 1 byte (+ 3 bytes de padding)
    C int32   // 4 bytes
}
// Taille totale : 12 bytes (pas 9 — à cause de l'alignement mémoire)
fmt.Println(unsafe.Sizeof(MaStruct{})) // 12

// Décalage d'un champ dans une struct
fmt.Println(unsafe.Offsetof(MaStruct{}.C)) // 8

// Conversion []byte → string sans copie (lecture seule !)
func bytesSansCopieSting(b []byte) string {
    return *(*string)(unsafe.Pointer(&b))
}

// Conversion string → []byte sans copie (DANGEREUX — ne pas modifier le résultat)
func stringSansCopieBytes(s string) []byte {
    return *(*[]byte)(unsafe.Pointer(&s))
}
```

> ⚠️ **`unsafe` : à utiliser avec une extrême prudence**
>
> - Le nom n'est pas ironique. `unsafe` contourne toutes les garanties de sécurité de Go.
> - Le code `unsafe` peut casser sans avertissement entre deux versions de Go.
> - Une erreur avec `unsafe` provoque des bugs mémoire aussi difficiles à déboguer qu'en C.
> - **Règle d'or** : n'utilisez `unsafe` que si les benchmarks prouvent que c'est nécessaire, et uniquement dans des fonctions isolées, bien testées, et bien documentées.
> - La grande majorité des programmes Go n'ont jamais besoin d'`unsafe`.

---

### Interaction avec le noyau — `syscall` et `golang.org/x/sys`

Pour interagir directement avec le noyau Linux, Go expose les syscalls via deux packages :

```go
import (
    "fmt"
    "syscall"
)

// Lire les statistiques système via syscall
func lireStatsSysteme() {
    var info syscall.Sysinfo_t
    if err := syscall.Sysinfo(&info); err != nil {
        fmt.Println("Erreur syscall :", err)
        return
    }

    fmt.Printf("Uptime          : %d secondes\n", info.Uptime)
    fmt.Printf("RAM totale      : %d Mo\n", info.Totalram/1024/1024)
    fmt.Printf("RAM disponible  : %d Mo\n", info.Freeram/1024/1024)
    fmt.Printf("Processus actifs: %d\n", info.Procs)
}

// Lire /proc/stat pour les stats CPU réelles
func lireCPUStat() (string, error) {
    données, err := os.ReadFile("/proc/stat")
    if err != nil {
        return "", fmt.Errorf("lecture /proc/stat : %w", err)
    }

    // La première ligne contient les stats CPU agrégées
    lignes := strings.Split(string(données), "\n")
    if len(lignes) > 0 {
        return lignes[0], nil
    }
    return "", nil
}
```

> 🔍 **`/proc` — Le système de fichiers virtuel Linux** — Sur Linux, `/proc` expose les informations du noyau sous forme de fichiers texte. C'est ainsi que `top`, `htop`, et tous les outils de monitoring lisent les métriques système. Go peut lire ces fichiers directement avec `os.ReadFile` — pas besoin de syscalls complexes pour la plupart des cas.

---

## 🛠️ Projet fil rouge — `gowatch` avec métriques CPU réelles

On remplace les valeurs simulées par de vraies métriques lues depuis le système.

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "runtime"
    "strconv"
    "strings"
    "time"
)

// StatCPU représente les compteurs CPU de /proc/stat
type StatCPU struct {
    User    uint64
    Nice    uint64
    System  uint64
    Idle    uint64
    IOWait  uint64
    IRQ     uint64
    SoftIRQ uint64
}

// lireStatCPU lit les statistiques CPU depuis /proc/stat
func lireStatCPU() (StatCPU, error) {
    f, err := os.Open("/proc/stat")
    if err != nil {
        return StatCPU{}, fmt.Errorf("lireStatCPU : %w", err)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        ligne := scanner.Text()
        if !strings.HasPrefix(ligne, "cpu ") {
            continue
        }

        champs := strings.Fields(ligne)
        if len(champs) < 8 {
            return StatCPU{}, fmt.Errorf("format /proc/stat inattendu")
        }

        var stat StatCPU
        valeurs := []*uint64{
            &stat.User, &stat.Nice, &stat.System, &stat.Idle,
            &stat.IOWait, &stat.IRQ, &stat.SoftIRQ,
        }
        for i, ptr := range valeurs {
            val, err := strconv.ParseUint(champs[i+1], 10, 64)
            if err != nil {
                return StatCPU{}, fmt.Errorf("parsing /proc/stat : %w", err)
            }
            *ptr = val
        }
        return stat, nil
    }
    return StatCPU{}, fmt.Errorf("/proc/stat : ligne cpu non trouvée")
}

// calculerUsageCPU mesure l'utilisation CPU sur un intervalle de temps
// En comparant deux snapshots de /proc/stat
func calculerUsageCPU(intervalle time.Duration) (float64, error) {
    avant, err := lireStatCPU()
    if err != nil {
        return 0, err
    }

    time.Sleep(intervalle)

    après, err := lireStatCPU()
    if err != nil {
        return 0, err
    }

    // Calculer les deltas
    idleDelta := float64((après.Idle + après.IOWait) - (avant.Idle + avant.IOWait))
    totalAvant := avant.User + avant.Nice + avant.System + avant.Idle +
        avant.IOWait + avant.IRQ + avant.SoftIRQ
    totalAprès := après.User + après.Nice + après.System + après.Idle +
        après.IOWait + après.IRQ + après.SoftIRQ
    totalDelta := float64(totalAprès - totalAvant)

    if totalDelta == 0 {
        return 0, nil
    }

    return (1 - idleDelta/totalDelta) * 100, nil
}

// lireRAM lit l'utilisation mémoire depuis /proc/meminfo
func lireRAM() (total, disponible uint64, err error) {
    f, err := os.Open("/proc/meminfo")
    if err != nil {
        return 0, 0, fmt.Errorf("lireRAM : %w", err)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        ligne := scanner.Text()
        champs := strings.Fields(ligne)
        if len(champs) < 2 {
            continue
        }
        valeur, err := strconv.ParseUint(champs[1], 10, 64)
        if err != nil {
            continue
        }
        switch champs[0] {
        case "MemTotal:":
            total = valeur * 1024 // kB → bytes
        case "MemAvailable:":
            disponible = valeur * 1024
        }
    }
    return total, disponible, scanner.Err()
}

func main() {
    fmt.Println("=== gowatch — Métriques système réelles ===\n")

    // Stats runtime Go
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)

    fmt.Printf("Runtime Go\n")
    fmt.Printf("  Version         : %s\n", runtime.Version())
    fmt.Printf("  CPU logiques    : %d\n", runtime.NumCPU())
    fmt.Printf("  Goroutines      : %d\n", runtime.NumGoroutine())
    fmt.Printf("  Heap allouée    : %.1f Mo\n", float64(memStats.HeapAlloc)/1024/1024)
    fmt.Printf("  Cycles GC       : %d\n\n", memStats.NumGC)

    // Stats CPU réelles (Linux uniquement)
    if runtime.GOOS == "linux" {
        fmt.Print("Calcul usage CPU (1 seconde)... ")
        usage, err := calculerUsageCPU(time.Second)
        if err != nil {
            fmt.Printf("Erreur : %v\n", err)
        } else {
            fmt.Printf("%.1f%%\n", usage)
        }

        // Stats RAM
        total, dispo, err := lireRAM()
        if err != nil {
            fmt.Printf("RAM : Erreur : %v\n", err)
        } else {
            utilisée := total - dispo
            pct := float64(utilisée) / float64(total) * 100
            fmt.Printf("RAM\n")
            fmt.Printf("  Totale          : %.1f Go\n", float64(total)/1024/1024/1024)
            fmt.Printf("  Utilisée        : %.1f Go (%.1f%%)\n",
                float64(utilisée)/1024/1024/1024, pct)
            fmt.Printf("  Disponible      : %.1f Go\n", float64(dispo)/1024/1024/1024)
        }
    } else {
        fmt.Printf("Note : métriques /proc disponibles uniquement sur Linux\n")
        fmt.Printf("OS courant : %s\n", runtime.GOOS)
    }
}
```

```bash
go run main.go
```

**Résultat sur Linux :**
```
=== gowatch — Métriques système réelles ===

Runtime Go
  Version         : go1.23.0
  CPU logiques    : 8
  Goroutines      : 1
  Heap allouée    : 0.3 Mo
  Cycles GC       : 0

Calcul usage CPU (1 seconde)... 12.3%
RAM
  Totale          : 15.9 Go
  Utilisée        : 9.8 Go (61.6%)
  Disponible      : 6.1 Go
```

---

## Bilan du Module 05

`gowatch` est maintenant un outil système complet :

| Fonctionnalité | Chapitre | Technique |
|----------------|----------|-----------|
| Scanner de ports concurrent | 5.1 | goroutines + sémaphore + Context |
| Export sécurisé TLS | 5.1 | `crypto/tls`, `net/http` |
| Chiffrement AES-GCM | 5.1 | `crypto/aes`, `crypto/cipher` |
| Métriques CPU réelles | 5.2 | `/proc/stat`, calcul de delta |
| Métriques RAM réelles | 5.2 | `/proc/meminfo` |
| Stats runtime Go | 5.2 | `runtime.MemStats` |

---

## Ce qu'il faut retenir

1. **Mesurer avant d'optimiser** — `go test -bench`, `-cpuprofile`, `-memprofile`. Pas d'intuition, que des données.

2. **Les allocations sont souvent le vrai problème** — `sync.Pool` pour réutiliser, `make([]T, 0, n)` pour pré-allouer, `strings.Builder` pour concaténer. Ces trois techniques couvrent 80% des optimisations Go.

3. **`runtime.MemStats`** — votre tableau de bord interne. Heap allouée, cycles GC, pauses. Indispensable pour comprendre ce que fait votre programme en mémoire.

4. **`unsafe` est le dernier recours** — pas le premier réflexe. Si vous en avez besoin, isolez-le, testez-le, documentez-le abondamment.

5. **`/proc` sur Linux** — le système de fichiers virtuel du noyau. Plus simple qu'un syscall pour la plupart des métriques système. Lisible avec `os.ReadFile` ou `bufio.Scanner`.

---

## Pour aller plus loin

- 📄 [Profiling Go Programs](https://go.dev/blog/pprof) — Le guide de référence
- 📄 [High Performance Go Workshop — Dave Cheney](https://dave.cheney.net/high-performance-go-workshop/dotgo-paris.html) — La référence sur l'optimisation Go
- 📄 [unsafe — Documentation officielle](https://pkg.go.dev/unsafe)
- 🔧 [goleak](https://github.com/uber-go/goleak) — Détecter les goroutine leaks dans les tests
- 🔧 [go-torch](https://github.com/uber-archive/go-torch) — Flamegraphs pour pprof
- 🔧 [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) — Comparer statistiquement deux séries de benchmarks

---

<div align="center">

[⬅️ Chapitre 5.1 — Réseau et Sécurité](./01-reseau-securite.md)

---

## 🎉 Félicitations — Vous avez terminé la formation

Vous avez parcouru un chemin complet : de la philosophie de Go jusqu'à ses recoins les plus techniques.

**Ce que vous savez maintenant :**
- Pourquoi Go existe et quand l'utiliser
- Écrire du code idiomatique Go — pas du "Java en Go"
- Maîtriser la concurrence : goroutines, channels, Context
- Construire des APIs robustes et les déployer en images Docker de 8 Mo
- Tester, benchmarker, et optimiser du code Go
- Interagir avec le réseau et le système au niveau bas niveau

**Les deux projets que vous avez construits :**
- `gowatch` — un outil CLI de monitoring système, concurrent, avec scanner réseau
- `gohub` — une API REST testée, dockerisée, prête pour la production

La suite ? Construisez quelque chose de réel avec Go. C'est le meilleur apprentissage qui reste.

[👉 Retour au README principal](../README.md)

</div>
