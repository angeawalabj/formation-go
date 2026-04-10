# Projets fil rouge

Ce dossier contient le code source complet des deux projets construits
tout au long de la formation.

---

## `gowatch` — Outil CLI de monitoring système

Un binaire unique, concurrent, sans dépendances.

```bash
cd gowatch
make build
./gowatch --watch
```

**Fonctionnalités :**
- Collecte concurrente des métriques (CPU, RAM, goroutines)
- Métriques système réelles via `/proc` sur Linux
- Mode surveillance avec rafraîchissement configurable
- Scanner de ports TCP concurrent
- Export HTTPS vers gohub
- Multi-plateforme (Linux, macOS, Windows, ARM)

---

## `gohub` — Dashboard API REST

Une API HTTP testée, dockerisée, en image de 8 Mo.

```bash
cd gohub
make build
./gohub
curl http://localhost:8080/health
```

**Fonctionnalités :**
- 4 endpoints REST (`/health`, `/api/metrics`, `/api/metrics/history`, `/api/metrics/ingest`)
- Store en mémoire thread-safe avec limite configurable
- Logs structurés JSON avec `log/slog`
- Graceful shutdown sur SIGINT/SIGTERM
- Tests unitaires + Table-Driven Tests + benchmarks
- Dockerfile multi-stage (image finale : scratch + binaire)

---

## Intégration complète

```bash
# Terminal 1 — démarrer gohub
cd gohub && ./gohub

# Terminal 2 — gowatch envoie ses métriques à gohub
cd gowatch && ./gowatch --watch --interval 2s \
    --export http://localhost:8080/api/metrics/ingest

# Terminal 3 — consulter le dashboard
watch -n 2 'curl -s http://localhost:8080/api/metrics | python3 -m json.tool'
```
