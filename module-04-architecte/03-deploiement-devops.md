# Chapitre 4.3 — Déploiement et DevOps

> *"Un binaire Go, c'est comme une valise qui contient déjà tout
> ce dont vous avez besoin pour le voyage."*

---

## Le problème

Votre API `gohub` fonctionne. Elle est testée. Elle est robuste.

Maintenant, il faut la livrer.

Dans la plupart des écosystèmes, déployer une application web ressemble à ceci :
- Installer le runtime (Node.js, Python, JVM) sur le serveur
- Copier les fichiers source
- Installer les dépendances (`npm install`, `pip install`, `mvn install`)
- Configurer les variables d'environnement
- Gérer les conflits de versions entre projets
- Maintenir tout ça à jour

Go transforme ce processus en quelque chose de radicalement différent.

Un binaire Go est **autonome**. Il contient tout ce dont il a besoin pour s'exécuter — le code, les dépendances, une partie du runtime. Vous le copiez sur n'importe quel serveur compatible et vous le lancez. C'est tout.

Ce chapitre vous montre comment aller de votre code source à une image Docker de moins de 10 Mo, déployable en une commande.

---

## L'intuition

### Le binaire unique comme philosophie de déploiement

La plupart des langages compilés produisent un exécutable qui dépend de bibliothèques système partagées (`libc`, `libssl`...). Ces bibliothèques doivent être présentes sur la machine cible, dans la bonne version.

Go peut produire un binaire **statiquement lié** — qui ne dépend de rien de ce qui est installé sur la machine cible. Pas de libc. Pas d'OpenSSL. Rien. Le binaire est entièrement autonome.

Ce n'est pas une petite optimisation — c'est une transformation de la façon dont on pense le déploiement.

### L'image Docker "Distroless"

Une image Docker standard Ubuntu ou Debian pèse plusieurs centaines de Mo. Elle contient un système d'exploitation complet — shell, utilitaires, gestionnaire de paquets, bibliothèques — dont votre application n'a besoin d'aucun.

Avec un binaire Go statique, votre image Docker peut ne contenir que... votre binaire. Rien d'autre. C'est ce qu'on appelle une image **distroless** ou **scratch** — une image vide dans laquelle on pose uniquement le binaire.

Résultat : des images de 5 à 15 Mo au lieu de 800 Mo. Des démarrages en millisecondes. Une surface d'attaque réduite à son minimum absolu.

---

## La solution Go

### Compilation — Les options essentielles

```bash
# Compilation simple — produit un binaire pour l'OS courant
go build -o gohub ./...

# Compilation avec optimisations pour la production
# -ldflags "-s -w" supprime les symboles de debug (-s) et la table DWARF (-w)
# Réduit la taille du binaire de 20 à 30%
go build -ldflags="-s -w" -o gohub ./...

# Injecter des métadonnées au moment de la compilation
# Permet d'avoir la version, le commit git, et la date dans le binaire
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build \
  -ldflags="-s -w \
    -X main.Version=${VERSION} \
    -X main.Commit=${COMMIT} \
    -X main.BuildDate=${DATE}" \
  -o gohub ./...
```

**Dans `main.go` — Variables injectées à la compilation :**
```go
package main

// Ces variables sont injectées par -ldflags au moment du build
// Leurs valeurs par défaut sont utilisées en développement local
var (
    Version   = "dev"
    Commit    = "unknown"
    BuildDate = "unknown"
)

func main() {
    fmt.Printf("gohub %s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
    // ...
}
```

```bash
./gohub --version
# gohub v1.2.3 (commit: a3f8c12, built: 2024-01-15T15:42:00Z)
```

> 💡 **`-X main.Version`** — Cette technique permet d'embarquer des métadonnées de build directement dans le binaire sans modifier le code source. C'est la façon idiomatique Go de versionner un binaire. Tous les outils CLI Go sérieux l'utilisent.

---

### Cross-compilation — Compiler pour n'importe quelle cible

C'est l'une des fonctionnalités les plus puissantes de Go : **compiler pour une autre plateforme depuis votre machine**.

Deux variables d'environnement contrôlent la cible :
- `GOOS` — le système d'exploitation cible
- `GOARCH` — l'architecture processeur cible

```bash
# Compiler pour Linux AMD64 (le serveur typique) depuis macOS ou Windows
GOOS=linux GOARCH=amd64 go build -o gohub-linux-amd64 ./...

# Compiler pour macOS ARM (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o gohub-darwin-arm64 ./...

# Compiler pour Windows
GOOS=windows GOARCH=amd64 go build -o gohub-windows.exe ./...

# Compiler pour Raspberry Pi (ARM 32 bits)
GOOS=linux GOARCH=arm GOARM=7 go build -o gohub-rpi ./...

# Compiler pour tous en une fois — script de release
for OS in linux darwin windows; do
    for ARCH in amd64 arm64; do
        EXT=""
        [ "$OS" = "windows" ] && EXT=".exe"
        GOOS=$OS GOARCH=$ARCH go build \
            -ldflags="-s -w" \
            -o "dist/gohub-${OS}-${ARCH}${EXT}" \
            ./...
        echo "Compilé : gohub-${OS}-${ARCH}${EXT}"
    done
done
```

> ⚠️ **CGo et la cross-compilation** — Si votre code utilise CGo (des bindings C), la cross-compilation devient beaucoup plus complexe. C'est une des raisons pour lesquelles la communauté Go évite CGo autant que possible. Pour `gohub`, on reste en Go pur — la cross-compilation fonctionne parfaitement.

---

### Binaire statique — Zéro dépendance système

Par défaut, Go peut utiliser CGo pour certaines opérations système (résolution DNS, par exemple). Pour produire un binaire véritablement statique :

```bash
# CGO_ENABLED=0 désactive complètement CGo
# Produit un binaire 100% statique, sans aucune dépendance externe
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o gohub \
    ./...

# Vérifier qu'il n'y a aucune dépendance dynamique
ldd gohub
# not a dynamic executable  ← Parfait
```

> 💡 **Pourquoi `CGO_ENABLED=0` ?** — Même sans code C dans votre projet, Go peut utiliser CGo pour des fonctions système (résolution DNS sur Linux, par exemple). Désactiver CGo force Go à utiliser ses implémentations Go pures de ces fonctions. Résultat : un binaire vraiment portable.

---

### Docker — De 800 Mo à 8 Mo

La technique du **multi-stage build** Docker est parfaite pour Go : on utilise une image complète pour compiler, et une image minimale pour exécuter.

**`Dockerfile` :**
```dockerfile
# ─── Stage 1 : Compilation ───────────────────────────────────────────────────
# Image complète avec Go installé — utilisée uniquement pour compiler
FROM golang:1.23-alpine AS builder

# Activer le module cache pour accélérer les builds répétés
WORKDIR /app

# Copier d'abord les fichiers de dépendances — optimise le cache Docker
# Si go.mod et go.sum ne changent pas, cette couche est mise en cache
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste du code source
COPY . .

# Compiler le binaire statique
# CGO_ENABLED=0 : binaire statique, aucune dépendance
# -s -w : réduire la taille du binaire
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /gohub \
    ./...

# ─── Stage 2 : Image finale ───────────────────────────────────────────────────
# scratch = image vide, absolument rien dedans
# Votre binaire est la seule chose dans cette image
FROM scratch

# Copier les certificats SSL depuis l'image builder
# Nécessaire si votre app fait des requêtes HTTPS externes
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copier uniquement le binaire compilé
COPY --from=builder /gohub /gohub

# Exposer le port
EXPOSE 8080

# Lancer le binaire directement — pas de shell, pas d'intermédiaire
ENTRYPOINT ["/gohub"]
```

**Construire et mesurer :**
```bash
# Construire l'image
docker build -t gohub:latest .

# Vérifier la taille
docker images gohub
# REPOSITORY   TAG       IMAGE ID       SIZE
# gohub        latest    a3f8c12b9e4d   8.2MB

# Comparer avec une image Go classique
docker images golang:1.23-alpine
# REPOSITORY       TAG           SIZE
# golang           1.23-alpine   232MB

# Lancer le container
docker run -p 8080:8080 gohub:latest

# Tester
curl http://localhost:8080/health
# {"status":"ok","version":"1.0.0","uptime":"2s"}
```

> 💡 **`scratch` vs `distroless`** — `scratch` est l'image la plus minimale possible (absolument vide). `distroless/static` de Google est légèrement plus grande mais inclut quelques utilitaires de sécurité. Pour une API Go pure sans CGo, `scratch` est parfait. Si vous avez besoin de déboguer dans le container, `distroless/static` est plus pratique.

---

### Logs structurés avec `log/slog`

En production, les logs doivent être **structurés** — parsables par des outils comme Elasticsearch, Datadog, ou CloudWatch. Go 1.21 a introduit `log/slog` dans la bibliothèque standard.

```go
import "log/slog"

// Configuration du logger — une fois au démarrage
func initLogger(format string) *slog.Logger {
    var handler slog.Handler

    if format == "json" {
        // Format JSON pour la production — parsable par les outils de log
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        // Format texte pour le développement — lisible par un humain
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }

    return slog.New(handler)
}

// Utilisation dans l'application
func main() {
    logger := initLogger("json")

    // Log avec attributs structurés
    logger.Info("serveur démarré",
        slog.String("addr", ":8080"),
        slog.String("version", Version),
        slog.String("env", "production"),
    )

    // Log d'erreur avec contexte
    logger.Error("connexion base de données échouée",
        slog.String("host", "db.example.com"),
        slog.Int("port", 5432),
        slog.String("erreur", err.Error()),
    )

    // Log avec durée — parfait pour les métriques de latence
    debut := time.Now()
    // ... traitement ...
    logger.Info("requête traitée",
        slog.String("method", "GET"),
        slog.String("path", "/api/metrics"),
        slog.Duration("latence", time.Since(debut)),
        slog.Int("status", 200),
    )
}
```

**Sortie JSON en production :**
```json
{"time":"2024-01-15T15:42:01Z","level":"INFO","msg":"serveur démarré","addr":":8080","version":"1.0.0","env":"production"}
{"time":"2024-01-15T15:42:03Z","level":"INFO","msg":"requête traitée","method":"GET","path":"/api/metrics","latence":"1.2ms","status":200}
```

**Sortie texte en développement :**
```
time=2024-01-15T15:42:01Z level=INFO msg="serveur démarré" addr=:8080 version=1.0.0
time=2024-01-15T15:42:03Z level=INFO msg="requête traitée" method=GET path=/api/metrics latence=1.2ms status=200
```

---

### Middleware de logging HTTP avec `slog`

On améliore le middleware de logging de `gohub` pour utiliser `slog` :

```go
// responseWriter capture le status code écrit par le handler
type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

// middlewareLoggingSlog — logging structuré de chaque requête HTTP
func middlewareLoggingSlog(logger *slog.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            debut := time.Now()
            rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

            next.ServeHTTP(rw, r)

            logger.Info("http request",
                slog.String("method",   r.Method),
                slog.String("path",     r.URL.Path),
                slog.Int("status",      rw.statusCode),
                slog.Duration("latence", time.Since(debut)),
                slog.String("ip",       r.RemoteAddr),
            )
        })
    }
}
```

---

### Profiling avec `pprof` — Trouver les goulots d'étranglement

`pprof` est l'outil de profiling intégré à Go. Il analyse la consommation CPU, mémoire, les goroutines bloquées, et plus encore — en production, sur un serveur live.

```go
import (
    _ "net/http/pprof"  // Import pour ses effets de bord — enregistre les handlers pprof
    "net/http"
)

func main() {
    // Exposer pprof sur un port séparé — jamais sur le port public !
    go func() {
        // Ce serveur ne doit être accessible qu'en interne
        http.ListenAndServe("localhost:6060", nil)
    }()

    // ... reste du programme ...
}
```

**Utiliser pprof :**
```bash
# Profiler le CPU pendant 30 secondes
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyser la mémoire allouée
go tool pprof http://localhost:6060/debug/pprof/heap

# Voir les goroutines actives — détecter les leaks
go tool pprof http://localhost:6060/debug/pprof/goroutine

# Ouvrir une interface web interactive (la plus utile)
go tool pprof -http=:8090 http://localhost:6060/debug/pprof/profile?seconds=10
# Puis ouvrez http://localhost:8090 dans votre navigateur
```

> ⚠️ **Sécurité** — N'exposez jamais `pprof` sur un port public. Il révèle l'intégralité de la structure interne de votre programme. Utilisez `localhost:6060` ou protégez-le par authentification.

---

## 🛠️ Projet fil rouge — `gohub` prêt pour la production

On assemble tout ce chapitre dans `gohub` : métadonnées de build, logs structurés, et Dockerfile de production.

**`main.go` final :**
```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    _ "net/http/pprof"
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
    // Initialiser le logger structuré
    logFormat := os.Getenv("LOG_FORMAT")
    if logFormat == "" {
        logFormat = "text"
    }

    var handler slog.Handler
    opts := &slog.HandlerOptions{Level: slog.LevelInfo}
    if logFormat == "json" {
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
        slog.Int("cpu_count",     runtime.NumCPU()),
    )

    // Initialiser le store
    s := store.New()

    // Snapshot initial
    s.Ajouter(store.Snapshot{
        Timestamp: time.Now(),
        Metriques: []store.Metrique{
            {Source: "CPU",        Valeur: float64(runtime.NumCPU()),       Unite: "cœurs"},
            {Source: "Goroutines", Valeur: float64(runtime.NumGoroutine()), Unite: "actives"},
        },
    })

    // Créer le serveur API
    srv := api.New(s, Version, logger)

    // Serveur pprof sur port séparé (localhost uniquement)
    go func() {
        logger.Info("pprof disponible", slog.String("addr", "localhost:6060"))
        http.ListenAndServe("localhost:6060", nil)
    }()

    // Port configurable via variable d'environnement
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    httpSrv := &http.Server{
        Addr:         ":" + port,
        Handler:      srv,
        ReadTimeout:  5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout:  120 * time.Second,
    }

    // Démarrer en arrière-plan
    go func() {
        logger.Info("serveur HTTP démarré", slog.String("addr", ":"+port))
        if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Error("erreur serveur HTTP", slog.String("erreur", err.Error()))
            os.Exit(1)
        }
    }()

    // Attendre un signal d'arrêt
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    sig := <-sigCh

    logger.Info("signal reçu, arrêt gracieux",
        slog.String("signal", sig.String()),
    )

    // Graceful shutdown — attendre que les requêtes en cours se terminent
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := httpSrv.Shutdown(ctx); err != nil {
        logger.Error("erreur pendant le shutdown", slog.String("erreur", err.Error()))
    }

    logger.Info("serveur arrêté proprement")
}
```

**`Dockerfile` de production :**
```dockerfile
# Stage 1 : Build
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildDate=${BUILD_DATE}" \
    -o /gohub \
    ./...

# Stage 2 : Image finale
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /gohub /gohub

ENV PORT=8080
ENV LOG_FORMAT=json

EXPOSE 8080

ENTRYPOINT ["/gohub"]
```

**`Makefile` — Automatiser les tâches courantes :**
```makefile
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: build test docker clean

build:
	go build -ldflags="-s -w \
		-X main.Version=$(VERSION) \
		-X main.Commit=$(COMMIT) \
		-X main.BuildDate=$(DATE)" \
		-o gohub ./...

test:
	go test -race -cover ./...

bench:
	go test -bench=. -benchmem ./...

docker:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(DATE) \
		-t gohub:$(VERSION) \
		-t gohub:latest \
		.

clean:
	rm -f gohub
	docker rmi gohub:latest 2>/dev/null || true
```

**Déploiement complet en 3 commandes :**
```bash
# 1. Tester
make test

# 2. Construire l'image Docker
make docker

# 3. Déployer
docker run -d \
    --name gohub \
    -p 8080:8080 \
    -e LOG_FORMAT=json \
    --restart unless-stopped \
    gohub:latest

# Vérifier
curl http://localhost:8080/health
docker logs gohub
```

**Résultat — logs JSON en production :**
```json
{"time":"2024-01-15T15:42:00Z","level":"INFO","msg":"démarrage gohub","version":"v1.2.3","commit":"a3f8c12","go_version":"go1.23.0","cpu_count":8}
{"time":"2024-01-15T15:42:00Z","level":"INFO","msg":"serveur HTTP démarré","addr":":8080"}
{"time":"2024-01-15T15:42:01Z","level":"INFO","msg":"http request","method":"GET","path":"/health","status":200,"latence":"0.4ms"}
```

---

## Bilan du Module 04

`gohub` est passé de zéro à un service de production complet :

| Étape | Chapitre | Résultat |
|-------|----------|---------|
| API REST fonctionnelle | 4.1 | 4 endpoints, store thread-safe, middleware |
| Suite de tests | 4.2 | Table-Driven Tests, httptest, benchmarks, 100% coverage store |
| Prêt pour la production | 4.3 | Logs structurés, Docker 8 Mo, graceful shutdown, pprof |

Un service Go de production en trois chapitres. C'est ça, la promesse de Go.

---

## Ce qu'il faut retenir

1. **`-ldflags="-s -w"`** — réduction de 20-30% de la taille du binaire sans impact sur les performances. À toujours utiliser en production.

2. **`-X main.Variable=valeur`** — injecter version, commit, date au moment du build. Un binaire doit savoir qui il est.

3. **`CGO_ENABLED=0`** — binaire statique, aucune dépendance système. La clé pour les images Docker minimales.

4. **Multi-stage Docker + `scratch`** — compiler dans une image complète, copier uniquement le binaire dans une image vide. De 800 Mo à 8 Mo.

5. **`log/slog`** — logs structurés intégrés depuis Go 1.21. JSON en production, texte en développement. Jamais de `fmt.Println` pour les logs.

6. **Graceful shutdown** — `httpSrv.Shutdown(ctx)` attend que les requêtes en cours se terminent avant de s'arrêter. Un service qui ne fait pas ça perd des données.

---

## Pour aller plus loin

- 📄 [Profiling Go Programs](https://go.dev/blog/pprof) — Le guide officiel sur pprof
- 📄 [log/slog — Documentation officielle](https://pkg.go.dev/log/slog)
- 📄 [Multi-stage builds — Documentation Docker](https://docs.docker.com/build/building/multi-stage/)
- 🔧 [goreleaser](https://goreleaser.com/) — Automatiser la release multi-plateforme de binaires Go
- 🔧 [ko](https://ko.build/) — Construire des images Docker Go sans Dockerfile

---

<div align="center">

[⬅️ Chapitre 4.2 — Tooling et Qualité Production](./02-tooling-qualite.md) · [👉 Module 05 — L'Homme de l'Ombre](../module-05-ombre/README.md)

</div>
