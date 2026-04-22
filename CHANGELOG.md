# Changelog

Toutes les modifications notables de cette formation sont documentées ici.

Le format est basé sur [Keep a Changelog](https://keepachangelog.com/fr/1.0.0/).

---

## [1.0.0] — 2024

### Ajouté

**Contenu de la formation**
- Module 01 — L'Éveil : historique de Go, philosophie, installation, premier binaire
- Module 02 — La Forge : variables, slices, maps, erreurs, structs, interfaces
- Module 03 — Le Super-Pouvoir : goroutines, channels, WaitGroup, Mutex, Context
- Module 04 — L'Architecte : API REST, tests Table-Driven, Docker multi-stage, slog
- Module 05 — L'Homme de l'Ombre (Bonus) : TCP/TLS, scanner de ports, AES-GCM, pprof
- Préface et Introduction avec guide par profil d'apprenant

**Projet `gowatch`**
- Collecte concurrente des métriques système (CPU, RAM, goroutines)
- Métriques réelles depuis `/proc` sur Linux
- Mode surveillance avec intervalle configurable (`--watch`)
- Scanner de ports TCP concurrent avec sémaphore (`--scan`)
- Export HTTPS vers gohub (`--export`)
- Build multi-plateforme : Linux, macOS, Windows, ARM
- Dockerfile multi-stage (image finale : scratch)
- Makefile complet avec `build`, `test`, `all-platforms`, `docker`

**Projet `gohub`**
- 4 endpoints REST : `/health`, `/api/metrics`, `/api/metrics/history`, `/api/metrics/ingest`
- Store en mémoire thread-safe avec `sync.RWMutex`
- Logs structurés JSON avec `log/slog`
- Graceful shutdown sur SIGINT/SIGTERM
- Suite de tests : Table-Driven Tests + `httptest` + benchmarks
- Dockerfile multi-stage (image finale : ~8 Mo)
- Makefile complet avec `build`, `test`, `bench`, `cover`, `docker`

**Infrastructure**
- CI GitHub Actions (`go test -race` sur push et pull request)
- Licence CC BY-NC-SA 4.0
- CONTRIBUTING.md

---

## À venir

- [ ] PDF ebook de la formation complète
- [ ] GitHub Pages avec la landing page
- [ ] Binaires pré-compilés dans les GitHub Releases
- [ ] Module 05 chapitre 5.1 : exemple TLS complet avec certificat auto-signé
- [ ] Tests pour le package `collector` de gowatch
- [ ] Tests pour le package `scanner` de gowatch