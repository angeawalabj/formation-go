# gohub

> Dashboard de monitoring via API REST — projet fil rouge de la formation Go

Une API HTTP qui stocke et expose les métriques collectées par `gowatch`.
Testée, dockerisée, prête pour la production en moins de 10 Mo.

---

## Démarrage rapide

```bash
# Depuis les sources
cd formation-go/projets/gohub
make build
./gohub

# Vérifier
curl http://localhost:8080/health
```

## Endpoints

| Méthode | URL | Description |
|---------|-----|-------------|
| `GET` | `/health` | Statut du service, version, uptime |
| `GET` | `/api/metrics` | Dernier snapshot collecté |
| `GET` | `/api/metrics/history?limit=N` | N derniers snapshots (max 100) |
| `POST` | `/api/metrics/ingest` | Ingérer un snapshot depuis gowatch |

## Exemples

```bash
# Healthcheck
curl http://localhost:8080/health
# {"status":"ok","version":"1.0.0","uptime":"5m30s"}

# Dernières métriques
curl http://localhost:8080/api/metrics

# Historique (10 derniers)
curl "http://localhost:8080/api/metrics/history?limit=10"

# Ingérer des métriques depuis gowatch
curl -X POST http://localhost:8080/api/metrics/ingest \
     -H "Content-Type: application/json" \
     -d '{"metriques":[{"source":"CPU","valeur":23.4,"unite":"%"}]}'
```

## Intégration avec gowatch

```bash
# gowatch envoie ses métriques à gohub toutes les 2 secondes
./gowatch --watch --interval 2s --export http://localhost:8080/api/metrics/ingest
```

## Options

```bash
./gohub --help

  --port        Port d'écoute HTTP (défaut: 8080)
  --debug-port  Port pprof localhost (défaut: 6060)
  --log-format  Format logs : text | json (défaut: text)
  --version     Afficher la version
```

## Tests

```bash
# Tests unitaires avec race detector
make test

# Rapport de couverture
make cover

# Benchmarks
make bench
```

## Docker

```bash
# Construire l'image (~8 Mo)
make docker

# Lancer
make docker-run

# Logs en production
docker logs gohub
```

## Structure

```
gohub/
├── main.go              ← Point d'entrée + graceful shutdown
├── api/
│   ├── server.go        ← Handlers HTTP + middleware logging
│   └── server_test.go   ← Tests des handlers avec httptest
├── store/
│   ├── store.go         ← Stockage en mémoire thread-safe
│   └── store_test.go    ← Table-Driven Tests + benchmarks
├── Makefile
└── Dockerfile
```

---

*Construit dans le cadre de la formation "[Apprendre Go : Pourquoi et Pour Quoi Faire ?](../../README.md)"*
