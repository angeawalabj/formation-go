# Chapitre 4.1 — Cloud Native et Microservices

> *"Le meilleur framework, c'est celui dont vous n'avez pas besoin."*

---

## Le problème

Quand un développeur venant de Node.js ou de Python veut créer un serveur web, son premier réflexe est d'installer un framework : Express, FastAPI, Django, Spring Boot.

Ce réflexe est compréhensible. Ces frameworks apportent du routing, de la gestion d'erreurs, du middleware, et des dizaines d'autres fonctionnalités en quelques lignes de configuration.

Mais ils apportent aussi quelque chose d'autre : de la **complexité cachée**. Des centaines de dépendances transitives. Des conventions implicites. Un comportement difficile à déboguer quand quelque chose ne marche pas comme prévu.

Go fait un choix différent. Sa bibliothèque standard inclut un serveur HTTP complet, performant, et prêt pour la production. Des entreprises comme Cloudflare et GitHub font tourner des millions de requêtes par seconde avec `net/http` pur — sans framework tiers.

Ce chapitre vous montre comment construire une API REST complète avec la seule bibliothèque standard.

---

## L'intuition

### Ce que `net/http` vous donne gratuitement

Le package `net/http` de Go n'est pas un micro-framework minimal. C'est une implémentation HTTP complète qui inclut :

- Un serveur concurrent par défaut — chaque requête dans sa propre goroutine
- Un client HTTP avec connection pooling
- Un routeur de base avec pattern matching
- Le support HTTPS natif
- Le streaming de réponses

Ce que vous devrez ajouter vous-même ou avec de petites bibliothèques :
- Un routeur avancé avec paramètres d'URL (`/users/:id`)
- Un middleware pipeline propre
- La validation des données entrantes

Pour 80% des APIs, `net/http` pur suffit. Pour les 20% restants, des bibliothèques légères comme `chi` ou `gorilla/mux` ajoutent le routing avancé sans le reste du poids d'un framework.

---

## La solution Go

### Le serveur HTTP minimal

```go
package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Bonjour depuis Go !")
    })

    fmt.Println("Serveur démarré sur :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        panic(err)
    }
}
```

```bash
go run main.go
curl http://localhost:8080/
# Bonjour depuis Go !
```

Onze lignes. Un serveur concurrent prêt pour la production. Chaque requête est automatiquement traitée dans sa propre goroutine — vous n'avez rien à configurer.

---

### Les handlers — Le cœur de `net/http`

Tout tourne autour de l'interface `http.Handler` :

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

N'importe quel type qui implémente `ServeHTTP` peut être un handler HTTP. C'est l'interface Go en action.

```go
func monHandler(w http.ResponseWriter, r *http.Request) {
    // Lire la méthode HTTP
    fmt.Println("Méthode :", r.Method)       // GET, POST, PUT, DELETE...

    // Lire l'URL et ses paramètres
    fmt.Println("URL :", r.URL.Path)          // /api/metrics
    fmt.Println("Query :", r.URL.Query())     // ?limit=10&page=2

    // Lire les headers
    fmt.Println("Content-Type :", r.Header.Get("Content-Type"))

    // Écrire un header de réponse — avant le body
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK) // 200

    fmt.Fprintln(w, `{"status": "ok"}`)
}
```

> ⚠️ **Ordre critique** — Appelez `w.Header().Set(...)` et `w.WriteHeader(...)` **avant** d'écrire le body. Une fois le body commencé, les headers sont figés.

---

### Réponses JSON — Le pattern standard

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"
)

// repondreJSON sérialise v en JSON et l'envoie avec le code status donné
func repondreJSON(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        http.Error(w, "Erreur d'encodage JSON", http.StatusInternalServerError)
    }
}

// repondreErreur envoie une réponse d'erreur JSON standardisée
func repondreErreur(w http.ResponseWriter, status int, message string) {
    repondreJSON(w, status, map[string]interface{}{
        "code":    status,
        "message": message,
    })
}

type MetriqueResponse struct {
    Timestamp  time.Time `json:"timestamp"`
    CPUCores   int       `json:"cpu_cores"`
    Goroutines int       `json:"goroutines"`
    Uptime     string    `json:"uptime"`
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        repondreErreur(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
        return
    }

    repondreJSON(w, http.StatusOK, MetriqueResponse{
        Timestamp:  time.Now(),
        CPUCores:   8,
        Goroutines: 12,
        Uptime:     "4h32m",
    })
}
```

> 💡 **`json.NewEncoder(w).Encode(v)` vs `json.Marshal`**
> `json.Marshal` encode en mémoire puis vous écrivez — deux étapes, une allocation.
> `json.NewEncoder(w).Encode(v)` écrit directement dans le ResponseWriter — une étape, aucune allocation intermédiaire. C'est la façon idiomatique en HTTP.

---

### Désérialiser du JSON entrant

```go
type CreerMetriqueRequest struct {
    Source string  `json:"source"`
    Valeur float64 `json:"valeur"`
    Unite  string  `json:"unite"`
}

func handleCreerMetrique(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        repondreErreur(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
        return
    }

    // Limiter la taille du body — protection contre les requêtes géantes
    r.Body = http.MaxBytesReader(w, r.Body, 1_048_576) // 1 Mo max

    var req CreerMetriqueRequest
    decoder := json.NewDecoder(r.Body)
    decoder.DisallowUnknownFields() // Rejeter les champs inconnus

    if err := decoder.Decode(&req); err != nil {
        repondreErreur(w, http.StatusBadRequest, "JSON invalide : "+err.Error())
        return
    }

    if req.Source == "" {
        repondreErreur(w, http.StatusBadRequest, "Le champ 'source' est requis")
        return
    }

    repondreJSON(w, http.StatusCreated, map[string]string{
        "message": "Métrique créée",
        "source":  req.Source,
    })
}
```

---

### Le Middleware — Comportement transversal sans modifier les handlers

Un middleware est une fonction qui prend un `http.Handler` et retourne un `http.Handler` :

```go
// Middleware de logging
func middlewareLogging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        debut := time.Now()
        next.ServeHTTP(w, r)
        fmt.Printf("[%s] %s %s — %v\n",
            time.Now().Format("15:04:05"),
            r.Method,
            r.URL.Path,
            time.Since(debut),
        )
    })
}

// Middleware de récupération des panics
func middlewareRecover(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                repondreErreur(w, http.StatusInternalServerError, "Erreur interne")
                fmt.Printf("PANIC récupéré : %v\n", err)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

// Chaîner les middlewares
func chainer(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}
```

> ⚠️ **Toujours configurer les timeouts** — `http.ListenAndServe` est pratique mais dangereux en production : pas de timeout. Utilisez `http.Server` avec des timeouts explicites :

```go
srv := &http.Server{
    Addr:         ":8080",
    Handler:      handler,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
```

---

### Connexion à une base de données SQL

```go
import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3" // Import pour ses effets de bord (enregistrement du driver)
)

func InitDB(chemin string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", chemin)
    if err != nil {
        return nil, fmt.Errorf("InitDB : %w", err)
    }

    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("InitDB ping : %w", err)
    }

    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    return db, nil
}

// Insertion avec placeholder — jamais de concaténation SQL
func InsererMetrique(db *sql.DB, source string, valeur float64, unite string) error {
    _, err := db.Exec(
        "INSERT INTO metriques (source, valeur, unite) VALUES (?, ?, ?)",
        source, valeur, unite,
    )
    return err
}

// Lecture avec gestion propre des rows
func LireMetriques(db *sql.DB, limite int) ([]MetriqueDB, error) {
    rows, err := db.Query(
        "SELECT id, source, valeur, unite FROM metriques ORDER BY id DESC LIMIT ?",
        limite,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close() // Toujours fermer

    var metriques []MetriqueDB
    for rows.Next() {
        var m MetriqueDB
        if err := rows.Scan(&m.ID, &m.Source, &m.Valeur, &m.Unite); err != nil {
            return nil, err
        }
        metriques = append(metriques, m)
    }
    return metriques, rows.Err() // Vérifier les erreurs d'itération
}
```

> 💡 **Placeholders** — Toujours `?` (SQLite, MySQL) ou `$1`, `$2` (PostgreSQL). Jamais de concaténation de strings SQL. C'est la protection de base contre les injections SQL.

---

## 🛠️ Projet fil rouge — Lancement de `gohub`

On initialise `gohub` — le second projet fil rouge.

```bash
mkdir gohub && cd gohub
go mod init github.com/votre-pseudo/gohub
```

**Structure :**
```
gohub/
├── main.go
├── api/
│   └── server.go     ← Handlers + routes + middlewares
└── store/
    └── memory.go     ← Stockage en mémoire thread-safe
```

**`store/memory.go` :**
```go
package store

import (
    "sync"
    "time"
)

type Metrique struct {
    Source string  `json:"source"`
    Valeur float64 `json:"valeur"`
    Unite  string  `json:"unite"`
}

type Snapshot struct {
    ID        int64      `json:"id"`
    Timestamp time.Time  `json:"timestamp"`
    Metriques []Metrique `json:"metriques"`
}

type Store struct {
    mu        sync.RWMutex
    snapshots []Snapshot
    nextID    int64
}

func New() *Store { return &Store{} }

func (s *Store) Ajouter(snap Snapshot) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.nextID++
    snap.ID = s.nextID
    s.snapshots = append(s.snapshots, snap)
    if len(s.snapshots) > 1000 {
        s.snapshots = s.snapshots[len(s.snapshots)-1000:]
    }
}

func (s *Store) Dernier() (Snapshot, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if len(s.snapshots) == 0 {
        return Snapshot{}, false
    }
    return s.snapshots[len(s.snapshots)-1], true
}

func (s *Store) Historique(limite int) []Snapshot {
    s.mu.RLock()
    defer s.mu.RUnlock()
    if limite > len(s.snapshots) {
        limite = len(s.snapshots)
    }
    result := make([]Snapshot, limite)
    copy(result, s.snapshots[len(s.snapshots)-limite:])
    return result
}
```

**`api/server.go` :**
```go
package api

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/votre-pseudo/gohub/store"
)

type Server struct {
    store     *store.Store
    startTime time.Time
    version   string
    mux       *http.ServeMux
}

func New(s *store.Store, version string) *Server {
    srv := &Server{
        store:     s,
        startTime: time.Now(),
        version:   version,
        mux:       http.NewServeMux(),
    }
    srv.registerRoutes()
    return srv
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Middleware de logging inline
    debut := time.Now()
    srv.mux.ServeHTTP(w, r)
    fmt.Printf("[%s] %s %s %v\n", time.Now().Format("15:04:05"), r.Method, r.URL.Path, time.Since(debut))
}

func (srv *Server) registerRoutes() {
    srv.mux.HandleFunc("/health",                srv.handleHealth)
    srv.mux.HandleFunc("/api/metrics",           srv.handleMetrics)
    srv.mux.HandleFunc("/api/metrics/history",   srv.handleHistory)
    srv.mux.HandleFunc("/api/metrics/ingest",    srv.handleIngest)
}

func (srv *Server) json(w http.ResponseWriter, status int, v interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(v)
}

func (srv *Server) erreur(w http.ResponseWriter, status int, msg string) {
    srv.json(w, status, map[string]interface{}{"code": status, "message": msg})
}

func (srv *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
    srv.json(w, http.StatusOK, map[string]string{
        "status":  "ok",
        "version": srv.version,
        "uptime":  time.Since(srv.startTime).Round(time.Second).String(),
    })
}

func (srv *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        srv.erreur(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
        return
    }
    snap, ok := srv.store.Dernier()
    if !ok {
        srv.erreur(w, http.StatusNotFound, "Aucune métrique disponible")
        return
    }
    srv.json(w, http.StatusOK, snap)
}

func (srv *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        srv.erreur(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
        return
    }
    limite := 10
    if l := r.URL.Query().Get("limit"); l != "" {
        if n, err := strconv.Atoi(l); err == nil && n > 0 && n <= 100 {
            limite = n
        }
    }
    srv.json(w, http.StatusOK, map[string]interface{}{
        "count":     len(srv.store.Historique(limite)),
        "snapshots": srv.store.Historique(limite),
    })
}

func (srv *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        srv.erreur(w, http.StatusMethodNotAllowed, "Méthode non autorisée")
        return
    }
    var snap store.Snapshot
    r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
    if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
        srv.erreur(w, http.StatusBadRequest, "JSON invalide : "+err.Error())
        return
    }
    snap.Timestamp = time.Now()
    srv.store.Ajouter(snap)
    srv.json(w, http.StatusCreated, map[string]string{"message": "Snapshot enregistré"})
}
```

**`main.go` :**
```go
package main

import (
    "fmt"
    "net/http"
    "runtime"
    "time"

    "github.com/votre-pseudo/gohub/api"
    "github.com/votre-pseudo/gohub/store"
)

func main() {
    s := store.New()

    // Données initiales
    s.Ajouter(store.Snapshot{
        Timestamp: time.Now(),
        Metriques: []store.Metrique{
            {Source: "CPU",        Valeur: float64(runtime.NumCPU()),      Unite: "cœurs"},
            {Source: "Goroutines", Valeur: float64(runtime.NumGoroutine()), Unite: "actives"},
        },
    })

    srv := api.New(s, "1.0.0")

    httpSrv := &http.Server{
        Addr:         ":8080",
        Handler:      srv,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    fmt.Println("gohub v1.0.0 démarré sur :8080")
    if err := httpSrv.ListenAndServe(); err != nil {
        panic(err)
    }
}
```

**Testez :**
```bash
go run ./...

curl http://localhost:8080/health
curl http://localhost:8080/api/metrics
curl http://localhost:8080/api/metrics/history?limit=5
curl -X POST http://localhost:8080/api/metrics/ingest \
     -H "Content-Type: application/json" \
     -d '{"metriques":[{"source":"test","valeur":42,"unite":"ms"}]}'
```

---

## Ce qu'il faut retenir

1. **`net/http` est production-ready** — pas besoin de framework. La bibliothèque standard Go est un serveur HTTP complet, concurrent, et performant.

2. **Toujours configurer les timeouts** — `http.Server` avec `ReadTimeout`, `WriteTimeout`, `IdleTimeout`. Un serveur sans timeout est vulnérable.

3. **`json.NewEncoder(w).Encode(v)`** — écriture directe dans le ResponseWriter. Pas d'allocation intermédiaire.

4. **Le middleware est `Handler → Handler`** — interface simple, chaînable, sans framework.

5. **Placeholders SQL toujours** — jamais de concaténation. `defer rows.Close()`. Vérifier `rows.Err()`.

---

## Pour aller plus loin

- 📄 [net/http — Documentation officielle](https://pkg.go.dev/net/http)
- 📄 [JSON and Go — Blog officiel Go](https://go.dev/blog/json)
- 🔧 [chi — Routeur léger compatible net/http](https://github.com/go-chi/chi)
- 🔧 [sqlx — Extension légère de database/sql](https://github.com/jmoiron/sqlx)

---

<div align="center">

[⬅️ Retour au Module 04](./README.md) · [👉 Chapitre 4.2 — Tooling et Qualité Production](./02-tooling-qualite.md)

</div>
