# Chapitre 3.3 — Patterns de concurrence avancés

> *"Le mutex est un aveu. Un aveu que votre design
> partage de l'état là où il ne devrait pas."*
> — Proverbe de la communauté Go
>
> *"Mais parfois, l'aveu est honnête et nécessaire."*
> — La réalité du terrain

---

## Le problème

Les channels sont élégants. Ils résolvent la majorité des problèmes de concurrence en Go.

Mais pas tous.

Imaginez un cache en mémoire partagé entre des dizaines de goroutines. Chaque goroutine lit et écrit des entrées indépendantes — pas les mêmes clés, pas les même valeurs. Passer chaque lecture et écriture par un channel serait non seulement verbeux, mais contre-productif : vous créeriez un goulot d'étranglement là où il n'y en avait pas.

Imaginez aussi un programme qui lance cent goroutines et doit attendre que **toutes** aient terminé avant de continuer. Compter manuellement les résultats dans un channel fonctionne — mais Go a un outil fait exactement pour ça.

Et imaginez une goroutine qui doit s'arrêter proprement quand l'utilisateur appuie sur Ctrl+C, ou quand une requête HTTP expire, ou quand le programme décide de l'annuler. Comment propager cette annulation à travers plusieurs niveaux de goroutines imbriquées ?

Ce chapitre couvre les trois outils qui répondent à ces situations : **`sync.WaitGroup`**, **`sync.Mutex`**, et **`context.Context`**.

---

## L'intuition

### WaitGroup — Le chef de chantier

Imaginez un chef de chantier qui envoie dix ouvriers travailler sur dix tâches différentes. Il ne peut pas partir avant qu'ils aient tous terminé. Mais il ne va pas non plus se tenir derrière chacun d'eux.

Il utilise un tableau de bord : quand un ouvrier commence, il coche une case. Quand il termine, il la décoche. Le chef attend que toutes les cases soient décochées.

C'est exactement ce que fait un `WaitGroup`.

### Mutex — Le verrou de vestiaire

Un seul verrou pour un casier partagé. Quand vous l'avez, personne d'autre ne peut ouvrir le casier. Quand vous avez fini, vous libérez le verrou et quelqu'un d'autre peut l'utiliser.

Simple, brutal, efficace. Et à utiliser avec parcimonie.

### Context — La télécommande d'arrêt

Imaginez une télécommande qui peut envoyer un signal d'arrêt à toutes les goroutines d'un programme simultanément — même celles imbriquées dans d'autres goroutines, même celles qui ne se connaissent pas entre elles.

C'est le `Context`. Il propage les annulations et les timeouts en cascade, de la goroutine parent jusqu'aux goroutines les plus profondes.

---

## La solution Go

### `sync.WaitGroup` — Synchroniser une flotte de goroutines

`WaitGroup` est le remplaçant propre du `time.Sleep` temporaire qu'on utilisait jusqu'ici.

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func traiter(id int, wg *sync.WaitGroup) {
    defer wg.Done()  // Signale "j'ai terminé" quand la fonction retourne
                     // defer garantit que Done() est toujours appelé,
                     // même si la fonction panique

    time.Sleep(time.Duration(id) * 10 * time.Millisecond)
    fmt.Printf("Tâche %d terminée\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1)           // +1 avant de lancer la goroutine
        go traiter(i, &wg)  // Passer par pointeur — WaitGroup ne doit pas être copié
    }

    wg.Wait()  // Bloque jusqu'à ce que le compteur atteigne 0
    fmt.Println("Toutes les tâches sont terminées.")
}
```

**Résultat :**
```
Tâche 1 terminée
Tâche 2 terminée
Tâche 3 terminée
Tâche 4 terminée
Tâche 5 terminée
Toutes les tâches sont terminées.
```

> ⚠️ **Les trois règles du WaitGroup**
>
> 1. **`wg.Add(1)` avant `go`** — jamais après. Si vous faites `go func()` puis `wg.Add(1)`, la goroutine peut terminer avant que `Add` soit appelé, et `Wait` se débloque trop tôt.
> 2. **`wg.Done()` dans un `defer`** — pour garantir qu'il est toujours appelé même en cas de panic ou de retour anticipé.
> 3. **Passer par pointeur `&wg`** — un `WaitGroup` copié est un `WaitGroup` cassé. Toujours passer l'adresse.

---

**WaitGroup + channel — Le combo gagnant**

Dans la pratique, `WaitGroup` et channels travaillent souvent ensemble :

```go
func collecterTout(sources []string) []string {
    resultats := make(chan string, len(sources))
    var wg sync.WaitGroup

    for _, source := range sources {
        wg.Add(1)
        s := source
        go func() {
            defer wg.Done()
            // Simuler une collecte
            time.Sleep(50 * time.Millisecond)
            resultats <- fmt.Sprintf("données de %s", s)
        }()
    }

    // Fermer le channel quand toutes les goroutines ont terminé
    go func() {
        wg.Wait()
        close(resultats)
    }()

    // Collecter tous les résultats
    var tous []string
    for r := range resultats {
        tous = append(tous, r)
    }
    return tous
}

func main() {
    sources := []string{"CPU", "RAM", "Disque", "Réseau"}
    resultats := collecterTout(sources)
    for _, r := range resultats {
        fmt.Println(r)
    }
}
```

Ce pattern — WaitGroup pour synchroniser, channel pour collecter, goroutine dédiée pour fermer — est un des plus utilisés en Go production.

---

### `sync.Mutex` — Protéger un état partagé

Quand plusieurs goroutines doivent lire **et écrire** la même variable, un `Mutex` (mutual exclusion) garantit qu'une seule goroutine à la fois accède à la variable critique.

```go
package main

import (
    "fmt"
    "sync"
)

// Compteur thread-safe grâce au Mutex
type CompteurSafe struct {
    mu    sync.Mutex  // Le verrou — convention : nommer mu
    valeur int
}

func (c *CompteurSafe) Incrementer() {
    c.mu.Lock()         // Verrouiller — une seule goroutine passe à la fois
    defer c.mu.Unlock() // Déverrouiller à la sortie — toujours avec defer

    c.valeur++
}

func (c *CompteurSafe) Lire() int {
    c.mu.Lock()
    defer c.mu.Unlock()

    return c.valeur
}

func main() {
    compteur := &CompteurSafe{}
    var wg sync.WaitGroup

    // 1000 goroutines incrémentent en même temps
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            compteur.Incrementer()
        }()
    }

    wg.Wait()
    fmt.Println("Compteur final :", compteur.Lire())
    // Toujours 1000 — jamais de race condition
}
```

**Vérifiez avec le race detector :**
```bash
go run -race main.go
# Aucun avertissement — le Mutex protège correctement
```

---

**`sync.RWMutex` — Optimiser les lectures fréquentes**

Quand les lectures sont bien plus fréquentes que les écritures, `RWMutex` permet à **plusieurs goroutines de lire simultanément**, mais impose une exclusivité totale pour les écritures.

```go
type CacheMetriques struct {
    mu      sync.RWMutex
    donnees map[string]float64
}

func (c *CacheMetriques) Ecrire(cle string, valeur float64) {
    c.mu.Lock()         // Exclusivité totale pour l'écriture
    defer c.mu.Unlock()
    c.donnees[cle] = valeur
}

func (c *CacheMetriques) Lire(cle string) (float64, bool) {
    c.mu.RLock()         // Plusieurs goroutines peuvent lire en même temps
    defer c.mu.RUnlock()
    valeur, ok := c.donnees[cle]
    return valeur, ok
}
```

> 💡 **Mutex vs Channel — Comment choisir ?**
>
> | Situation | Outil recommandé |
> |-----------|-----------------|
> | Transférer des données entre goroutines | Channel |
> | Distribuer du travail | Channel |
> | Signaler qu'une tâche est terminée | Channel ou WaitGroup |
> | Protéger un état partagé (cache, compteur) | Mutex |
> | Lecture fréquente, écriture rare | RWMutex |
>
> En cas de doute, commencez par les channels. Si vous vous retrouvez à écrire des patterns complexes de channels pour protéger une variable simple, considérez le Mutex.

---

### `sync.Map` — La map thread-safe intégrée

Go fournit une map thread-safe dans le package `sync` : `sync.Map`. Elle est optimisée pour deux cas d'usage spécifiques :

1. Chaque clé n'est écrite qu'une seule fois, mais lue de nombreuses fois
2. Des goroutines différentes lisent et écrivent des clés différentes (pas les mêmes)

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var cache sync.Map

    // Écrire en parallèle depuis plusieurs goroutines
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            cle := fmt.Sprintf("source_%d", id)
            cache.Store(cle, float64(id)*10.5)
        }(i)
    }
    wg.Wait()

    // Lire toutes les entrées
    cache.Range(func(cle, valeur interface{}) bool {
        fmt.Printf("%s : %.1f\n", cle, valeur)
        return true  // Retourner false arrête l'itération
    })

    // Lire une entrée spécifique
    if val, ok := cache.Load("source_3"); ok {
        fmt.Println("source_3 :", val)
    }

    // Supprimer une entrée
    cache.Delete("source_0")
}
```

> ⚠️ **`sync.Map` n'est pas une map universelle** — pour un usage général avec des lectures et écritures mélangées sur les mêmes clés, une `map` normale avec un `RWMutex` est souvent plus performante. `sync.Map` brille uniquement dans ses deux cas d'usage spécifiques.

---

### `context.Context` — Le chef d'orchestre des annulations

Le `Context` est l'outil le plus important de ce chapitre — et probablement le plus sous-estimé par les débutants.

Il résout un problème fondamental : **comment arrêter proprement une goroutine depuis l'extérieur ?**

Une goroutine ne peut pas être tuée de force en Go. Elle doit accepter de s'arrêter. Le `Context` est le mécanisme qui lui transmet ce signal d'arrêt.

---

**Création d'un Context**

```go
import "context"

// Context de base — ne s'annule jamais seul
ctx := context.Background()

// Context avec annulation manuelle
ctx, cancel := context.WithCancel(context.Background())
defer cancel()  // Toujours appeler cancel() — libère les ressources

// Context avec timeout — s'annule automatiquement après la durée
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Context avec deadline — s'annule à une heure précise
deadline := time.Now().Add(5 * time.Second)
ctx, cancel := context.WithDeadline(context.Background(), deadline)
defer cancel()
```

---

**Utiliser un Context dans une goroutine**

```go
func travaillerAvecContext(ctx context.Context, id int) error {
    for {
        select {
        case <-ctx.Done():
            // Le contexte a été annulé — arrêt propre
            return fmt.Errorf("tâche %d annulée : %w", id, ctx.Err())

        default:
            // Continuer le travail normal
            fmt.Printf("Tâche %d : en cours...\n", id)
            time.Sleep(500 * time.Millisecond)
        }
    }
}

func main() {
    // Context avec timeout de 2 secondes
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    var wg sync.WaitGroup

    for i := 1; i <= 3; i++ {
        wg.Add(1)
        id := i
        go func() {
            defer wg.Done()
            if err := travaillerAvecContext(ctx, id); err != nil {
                fmt.Printf("Tâche %d terminée : %v\n", id, err)
            }
        }()
    }

    wg.Wait()
    fmt.Println("Toutes les tâches ont été arrêtées.")
}
```

**Résultat :**
```
Tâche 1 : en cours...
Tâche 2 : en cours...
Tâche 3 : en cours...
Tâche 1 : en cours...
Tâche 2 : en cours...
Tâche 3 : en cours...
Tâche 1 : en cours...
Tâche 2 : en cours...
Tâche 3 : en cours...
Tâche 1 : en cours...
Tâche 2 : en cours...
Tâche 3 : en cours...
Tâche 1 terminée : tâche 1 annulée : context deadline exceeded
Tâche 2 terminée : tâche 2 annulée : context deadline exceeded
Tâche 3 terminée : tâche 3 annulée : context deadline exceeded
Toutes les tâches ont été arrêtées.
```

---

**Propagation du Context en cascade**

Le vrai pouvoir du Context est sa **propagation** : un Context annulé annule automatiquement tous ses enfants.

```go
func collecterAvecCtx(ctx context.Context, source string) (float64, error) {
    // Simuler une collecte lente
    select {
    case <-time.After(100 * time.Millisecond):
        return 42.0, nil
    case <-ctx.Done():
        return 0, fmt.Errorf("collecte %s annulée : %w", source, ctx.Err())
    }
}

func orchestrer(ctx context.Context) error {
    sources := []string{"CPU", "RAM", "Disque"}
    resultats := make(chan float64, len(sources))
    erreurs := make(chan error, len(sources))

    var wg sync.WaitGroup

    for _, src := range sources {
        wg.Add(1)
        s := src
        go func() {
            defer wg.Done()
            // Chaque goroutine reçoit le même context parent
            val, err := collecterAvecCtx(ctx, s)
            if err != nil {
                erreurs <- err
                return
            }
            resultats <- val
        }()
    }

    // Fermer les channels quand tout est terminé
    go func() {
        wg.Wait()
        close(resultats)
        close(erreurs)
    }()

    // Vérifier les erreurs
    for err := range erreurs {
        return err  // Première erreur = on remonte
    }

    // Afficher les résultats
    for r := range resultats {
        fmt.Printf("Résultat : %.2f\n", r)
    }
    return nil
}

func main() {
    // Timeout global de 50ms — trop court pour les collectes de 100ms
    ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
    defer cancel()

    if err := orchestrer(ctx); err != nil {
        fmt.Println("Erreur :", err)
    }
}
```

---

**Passer des valeurs dans un Context**

Le Context peut aussi transporter des valeurs métier à travers les couches de votre application :

```go
type cleContexte string  // Type custom pour éviter les collisions de clés

const (
    CleRequestID cleContexte = "request_id"
    CleUserID    cleContexte = "user_id"
)

func traiterRequete(ctx context.Context) {
    // Lire une valeur depuis le context
    requestID := ctx.Value(CleRequestID)
    fmt.Printf("Traitement de la requête %v\n", requestID)
}

func main() {
    // Injecter des valeurs dans le context
    ctx := context.WithValue(context.Background(), CleRequestID, "REQ-2024-001")
    ctx  = context.WithValue(ctx, CleUserID, "user-42")

    traiterRequete(ctx)
}
```

> ⚠️ **Utilisez `context.WithValue` avec parcimonie** — il est conçu pour des données transversales (request ID, user ID, trace ID), pas pour passer des paramètres de fonction déguisés. Si une donnée est nécessaire à une fonction, passez-la en paramètre explicite. Le Context n'est pas un fourre-tout.

---

## 🛠️ Projet fil rouge — `gowatch` avec Context et arrêt propre

On intègre le `Context` dans `gowatch` pour une gestion propre des timeouts et de l'arrêt.

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "runtime"
    "sync"
    "syscall"
    "time"
)

type Metrique struct {
    Source string
    Valeur float64
    Unite  string
    Erreur error
}

type Source struct {
    Nom      string
    Unite    string
    Timeout  time.Duration
    Collecte func(ctx context.Context) (float64, error)
}

// collecterAvecTimeout collecte une source avec son propre timeout
func collecterAvecTimeout(ctx context.Context, src Source, out chan<- Metrique) {
    // Context enfant avec timeout spécifique à cette source
    srcCtx, cancel := context.WithTimeout(ctx, src.Timeout)
    defer cancel()

    done := make(chan struct {
        val float64
        err error
    }, 1)

    go func() {
        val, err := src.Collecte(srcCtx)
        done <- struct {
            val float64
            err error
        }{val, err}
    }()

    select {
    case result := <-done:
        out <- Metrique{
            Source: src.Nom,
            Valeur: result.val,
            Unite:  src.Unite,
            Erreur: result.err,
        }
    case <-srcCtx.Done():
        out <- Metrique{
            Source: src.Nom,
            Erreur: fmt.Errorf("timeout après %v", src.Timeout),
        }
    }
}

// collecterSnapshot collecte toutes les sources en parallèle
func collecterSnapshot(ctx context.Context, sources []Source) []Metrique {
    out := make(chan Metrique, len(sources))
    var wg sync.WaitGroup

    for _, src := range sources {
        wg.Add(1)
        s := src
        go func() {
            defer wg.Done()
            collecterAvecTimeout(ctx, s, out)
        }()
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    var metriques []Metrique
    for m := range out {
        metriques = append(metriques, m)
    }
    return metriques
}

// afficherSnapshot affiche les métriques collectées
func afficherSnapshot(metriques []Metrique, duree time.Duration) {
    fmt.Printf("[%s] — collecté en %v\n", time.Now().Format("15:04:05"), duree)
    for _, m := range metriques {
        if m.Erreur != nil {
            fmt.Printf("  ✗ %-15s : %v\n", m.Source, m.Erreur)
        } else {
            fmt.Printf("  ✓ %-15s : %.0f %s\n", m.Source, m.Valeur, m.Unite)
        }
    }
    fmt.Println()
}

func main() {
    sources := []Source{
        {
            Nom:     "CPU Logiques",
            Unite:   "cœurs",
            Timeout: 100 * time.Millisecond,
            Collecte: func(ctx context.Context) (float64, error) {
                select {
                case <-time.After(10 * time.Millisecond):
                    return float64(runtime.NumCPU()), nil
                case <-ctx.Done():
                    return 0, ctx.Err()
                }
            },
        },
        {
            Nom:     "Goroutines",
            Unite:   "actives",
            Timeout: 100 * time.Millisecond,
            Collecte: func(ctx context.Context) (float64, error) {
                select {
                case <-time.After(5 * time.Millisecond):
                    return float64(runtime.NumGoroutine()), nil
                case <-ctx.Done():
                    return 0, ctx.Err()
                }
            },
        },
        {
            Nom:     "Source Lente",  // Simule une source qui dépasse son timeout
            Unite:   "ms",
            Timeout: 30 * time.Millisecond,
            Collecte: func(ctx context.Context) (float64, error) {
                select {
                case <-time.After(200 * time.Millisecond): // Trop lent
                    return 42, nil
                case <-ctx.Done():
                    return 0, ctx.Err()
                }
            },
        },
    }

    // Context racine — annulé sur Ctrl+C
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Écouter les signaux système
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh
        fmt.Println("\nSignal reçu — annulation en cours...")
        cancel()  // Annule le context racine — propage à toutes les goroutines
    }()

    fmt.Println("=== gowatch v0.7 — Context + Timeouts par source ===\n")

    ticker := time.NewTicker(3 * time.Second)
    defer ticker.Stop()

    snapshots := 0

    for {
        select {
        case <-ticker.C:
            snapshots++
            debut := time.Now()
            metriques := collecterSnapshot(ctx, sources)
            afficherSnapshot(metriques, time.Since(debut))

        case <-ctx.Done():
            fmt.Printf("Arrêt propre. %d snapshot(s) collecté(s).\n", snapshots)
            return
        }
    }
}
```

**Lancez et observez :**
```bash
go run main.go
```

**Résultat :**
```
=== gowatch v0.7 — Context + Timeouts par source ===

[15:42:01] — collecté en 31ms
  ✓ CPU Logiques    : 8 cœurs
  ✓ Goroutines      : 7 actives
  ✗ Source Lente    : timeout après 30ms

[15:42:04] — collecté en 31ms
  ✓ CPU Logiques    : 8 cœurs
  ✓ Goroutines      : 7 actives
  ✗ Source Lente    : timeout après 30ms

^C
Signal reçu — annulation en cours...
Arrêt propre. 2 snapshot(s) collecté(s).
```

`gowatch` est maintenant un outil robuste : chaque source a son propre timeout, les sources lentes n'impactent pas les sources rapides, et l'arrêt est propre à tous les niveaux.

---

## Ce qu'il faut retenir

1. **`sync.WaitGroup`** — l'outil pour attendre un groupe de goroutines. `Add(1)` avant `go`, `Done()` dans un `defer`, `Wait()` pour bloquer. Toujours passer par pointeur.

2. **`sync.Mutex`** — protège un état partagé quand les channels ne sont pas adaptés. `Lock()` / `Unlock()` toujours en paire, `Unlock()` toujours dans un `defer`. `RWMutex` pour optimiser les lectures fréquentes.

3. **`context.Context`** — propage les annulations et les timeouts en cascade. Toujours accepter un `ctx` en premier paramètre dans les fonctions qui font du travail long. Toujours appeler `cancel()` dans un `defer`.

4. **La hiérarchie Context** — un context annulé annule automatiquement tous ses enfants. C'est ce qui permet d'annuler proprement une arborescence entière de goroutines avec un seul `cancel()`.

5. **Le combo gagnant** — `WaitGroup` pour synchroniser + `channel` pour collecter + `Context` pour annuler. Ces trois outils ensemble couvrent 95% des besoins de concurrence en production.

---

## Bilan du Module 03

Vous avez maintenant tous les outils de la concurrence Go :

| Problème | Outil |
|----------|-------|
| Lancer une tâche en arrière-plan | `go maFonction()` |
| Faire passer des données entre goroutines | `channel` |
| Attendre plusieurs goroutines | `sync.WaitGroup` |
| Protéger un état partagé | `sync.Mutex` / `sync.RWMutex` |
| Map thread-safe | `sync.Map` |
| Propager une annulation / un timeout | `context.Context` |
| Attendre le premier channel prêt | `select` |
| Détecter les race conditions | `go run -race` |

`gowatch` est passé d'un programme séquentiel à un daemon concurrent, robuste, avec gestion des timeouts par source et arrêt propre sur signal. C'est du Go de production.

---

## Pour aller plus loin

- 📄 [Go Concurrency Patterns: Context](https://go.dev/blog/context) — L'article officiel sur le Context API
- 📄 [The sync package](https://pkg.go.dev/sync) — Documentation officielle de sync
- 📄 [Rethinking Classical Concurrency Patterns](https://drive.google.com/file/d/1nPdvhB0PutEJzdCq5ms6UI58dp50fcAN/view) — Bryan Mills, GopherCon 2018
- 🔧 [goleak](https://github.com/uber-go/goleak) — L'outil d'Uber pour détecter les goroutine leaks dans les tests

---

<div align="center">

[⬅️ Chapitre 3.2 — Channels](./02-channels.md) · [👉 Module 04 — L'Architecte](../module-04-architecte/README.md)

</div>
