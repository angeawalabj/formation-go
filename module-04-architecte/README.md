# Module 04 — L'Architecte

> *"Un programme qui marche sur votre machine ne vaut rien.
> Un programme qui marche sur toutes les machines — c'est ça, l'ingénierie."*

---

## Ce que ce module va changer

Les trois premiers modules vous ont donné les fondations. Vous savez penser en Go, écrire du code idiomatique, et gérer la concurrence avec confiance.

Maintenant, on construit des **systèmes réels**.

Ce module aborde Go tel qu'il est utilisé dans les entreprises qui font tourner des millions de requêtes par jour. Pas du code de démonstration — du code qu'on déploie, qu'on teste, qu'on monitore, et qu'on maintient dans le temps.

C'est aussi le module où le second projet fil rouge entre en scène : **`gohub`**, le dashboard de monitoring via API REST.

---

## Les objectifs de ce module

À la fin du Module 04, vous serez capable de :

- ✅ Créer un serveur HTTP performant avec la bibliothèque standard Go
- ✅ Structurer une API REST avec routing, middleware et gestion d'erreurs
- ✅ Sérialiser et désérialiser du JSON de façon robuste
- ✅ Connecter votre application à une base de données SQL
- ✅ Écrire des tests unitaires idiomatiques avec Table-Driven Tests
- ✅ Mesurer les performances avec des benchmarks
- ✅ Compiler un binaire Go pour n'importe quelle cible
- ✅ Dockeriser une application Go en image de moins de 10 Mo
- ✅ Observer votre application en production : logs, métriques, profiling

---

## Ce que vous allez construire

> 🛠️ **Projet fil rouge — `gohub` : Le dashboard API REST**

`gohub` est une API HTTP qui expose les métriques collectées par `gowatch`.
Elle démarre dans ce module et sera enrichie jusqu'au déploiement final.

```bash
# Démarrer le serveur
$ ./gohub --port 8080

# Consulter les métriques en temps réel
$ curl http://localhost:8080/api/metrics
{
  "timestamp": "2024-01-15T15:42:01Z",
  "cpu_cores": 8,
  "goroutines": 12,
  "uptime": "4h32m15s"
}

# Historique des 10 derniers snapshots
$ curl http://localhost:8080/api/metrics/history?limit=10

# Healthcheck — pour les load balancers et orchestrateurs
$ curl http://localhost:8080/health
{"status": "ok", "version": "1.0.0"}
```

---

## Les chapitres de ce module

### [Chapitre 4.1 — Cloud Native et Microservices](./01-cloud-native-microservices.md)
Construire une API REST complète avec la bibliothèque standard Go — sans framework externe. Routing, middleware, JSON, Context HTTP, connexion base de données.

**Concepts abordés :** `net/http`, handlers, middleware, `encoding/json`, `database/sql`.

---

### [Chapitre 4.2 — Tooling et Qualité Production](./02-tooling-qualite.md)
Tester et mesurer son code à la façon Go. Table-Driven Tests, mocking avec interfaces, benchmarks, et gestion des dépendances avec Go Modules.

**Concepts abordés :** `testing`, Table-Driven Tests, benchmarks, `go test`, `go mod`.

---

### [Chapitre 4.3 — Déploiement et DevOps](./03-deploiement-devops.md)
Du binaire compilé à l'image Docker de 5 Mo. Compilation croisée, multi-stage builds, logs structurés, et profiling avec `pprof`.

**Concepts abordés :** cross-compilation, Docker multi-stage, `log/slog`, `pprof`.

---

## Durée estimée

| Chapitre | Lecture | Pratique | Total |
|----------|---------|----------|-------|
| 4.1 — Cloud Native et Microservices | 35 min | 40 min | 75 min |
| 4.2 — Tooling et Qualité Production | 30 min | 35 min | 65 min |
| 4.3 — Déploiement et DevOps | 25 min | 30 min | 55 min |
| **Total Module 04** | **90 min** | **105 min** | **~3h15** |

---

## L'état d'esprit à adopter

Ce module marque un changement de registre.

On ne parle plus de concepts à comprendre — on parle de **décisions d'ingénierie** à prendre. Pourquoi ce routeur plutôt qu'un autre ? Pourquoi ces tests plutôt que ceux-là ? Pourquoi cette stratégie de déploiement ?

Go est opiniatre sur beaucoup de choses — mais il laisse une liberté réelle sur l'architecture. Ce module vous donnera les outils pour faire des choix éclairés, pas des recettes à copier.

---

<div align="center">

[⬅️ Retour au Module 03](../module-03-super-pouvoir/README.md) · [👉 Commencer le Chapitre 4.1](./01-cloud-native-microservices.md)

</div>
