# gowatch

> Outil CLI de monitoring système — projet fil rouge de la formation Go

Un binaire unique, concurrent, sans dépendances externes.

---

## Installation

```bash
# Depuis les sources
git clone https://github.com/votre-pseudo/formation-go
cd formation-go/projets/gowatch
make build

# Vérifier
./gowatch --version
```

## Utilisation

```bash
# Snapshot unique (texte)
./gowatch

# Snapshot en JSON
./gowatch --format json

# Mode surveillance — rafraîchissement toutes les 2 secondes
./gowatch --watch
./gowatch --watch --interval 5s

# Scanner les ports d'un hôte
./gowatch --scan localhost --ports 1-1024
./gowatch --scan 192.168.1.1 --ports 22-443

# Exporter vers gohub
./gowatch --export http://localhost:8080/api/metrics/ingest
./gowatch --export https://gohub.example.com/api/metrics/ingest

# Combinaisons
./gowatch --watch --format json --export http://localhost:8080/api/metrics/ingest
```

## Exemple de sortie

```
=== gowatch · 2024-01-15T15:42:01Z ===
  OS      : linux/amd64
  Go      : go1.23.0

Métriques :
  ✓ CPU Logiques      : 8.00 cœurs
  ✓ Goroutines        : 4.00 actives
  ✓ Heap Go           : 0.30 Mo
  ✓ CPU Usage         : 12.30 %
  ✓ RAM Utilisée      : 61.20 %
  ✓ RAM Disponible    : 6144.00 Mo
```

## Structure

```
gowatch/
├── main.go               ← Point d'entrée CLI + flags
├── collector/
│   └── collector.go      ← Collecte concurrente des métriques
├── renderer/
│   └── renderer.go       ← Affichage texte et JSON
├── scanner/
│   └── scanner.go        ← Scanner de ports TCP concurrent
├── exporter/
│   └── exporter.go       ← Export HTTPS vers gohub
├── Makefile
└── Dockerfile
```

## Build multi-plateforme

```bash
make all-platforms
# dist/gowatch-linux-amd64
# dist/gowatch-linux-arm64
# dist/gowatch-darwin-amd64
# dist/gowatch-darwin-arm64
# dist/gowatch-windows.exe
```

## Docker

```bash
docker build -t gowatch:latest .
docker run gowatch:latest --format json
```

---

*Construit dans le cadre de la formation "[Apprendre Go : Pourquoi et Pour Quoi Faire ?](../../README.md)"*
