# Module 05 — L'Homme de l'Ombre ⭐ Bonus

> *"Go est le seul langage moderne où écrire un scanner de ports,
> une API REST, et un outil de cryptographie demande
> exactement le même niveau d'effort."*

---

## Ce que ce module est — et ce qu'il n'est pas

Ce module est **indépendant**. Vous pouvez l'aborder après le Module 03 si votre objectif est purement système, ou après le Module 04 pour une vision complète.

Il n'est pas obligatoire pour construire des APIs ou des outils CLI classiques. Mais si vous voulez comprendre Go dans ses retranchements — comment il interagit avec le réseau bas niveau, comment il gère la cryptographie, comment on optimise un programme Go au-delà du raisonnable — c'est ici.

Ce module cible principalement trois profils :
- **DevOps / SRE** — qui veulent écrire des outils d'infrastructure réseau
- **Ingénieurs sécurité** — qui veulent comprendre les outils offensifs et défensifs
- **Développeurs curieux** — qui veulent aller au bout de ce que Go peut faire

---

## Les objectifs de ce module

À la fin du Module 05, vous serez capable de :

- ✅ Écrire un scanner de ports concurrent et configurable
- ✅ Créer des connexions TCP/UDP bas niveau
- ✅ Implémenter du tunneling TCP simple
- ✅ Manipuler les certificats TLS et faire du HTTPS natif
- ✅ Utiliser les primitives cryptographiques de Go (hashing, chiffrement)
- ✅ Profiler et optimiser un programme Go avec `pprof`
- ✅ Comprendre quand et comment utiliser `unsafe`
- ✅ Interagir avec le système Linux via les syscalls Go

---

## Ce que vous allez construire

> 🛠️ **Projet fil rouge — `gowatch` Version Pro**

La version finale de `gowatch` intègre :

```bash
# Scanner les ports ouverts de l'hôte local
$ ./gowatch --scan localhost --ports 1-1024
[15:42:01] Scan de localhost (ports 1-1024) — concurrent
  ✓ Port 22   : ouvert (SSH)
  ✓ Port 80   : ouvert (HTTP)
  ✓ Port 443  : ouvert (HTTPS)
  ✓ Port 8080 : ouvert (gohub)
  Scan terminé en 1.2s (1024 ports testés)

# Export des métriques chiffré via TLS
$ ./gowatch --export https://gohub.local:8443/api/metrics/ingest \
            --cert client.crt --key client.key

# Profiling embarqué activable à la demande
$ ./gowatch --profile cpu --duration 30s
CPU profile sauvegardé dans gowatch-cpu-20240115.prof
```

---

## Les chapitres de ce module

### [Chapitre 5.1 — Réseau et Sécurité Système](./01-reseau-securite.md)
TCP bas niveau, scanner de ports concurrent, tunneling, TLS natif, et cryptographie Go.

**Concepts abordés :** `net`, `crypto/tls`, `crypto/sha256`, `crypto/aes`, scanner de ports.

---

### [Chapitre 5.2 — Optimisation Bas Niveau](./02-optimisation-bas-niveau.md)
Profiling avancé, réduction des allocations mémoire, `unsafe`, et interaction avec le noyau Linux.

**Concepts abordés :** `pprof`, `runtime`, `unsafe`, `syscall`, optimisation mémoire.

---

## Durée estimée

| Chapitre | Lecture | Pratique | Total |
|----------|---------|----------|-------|
| 5.1 — Réseau et Sécurité | 30 min | 35 min | 65 min |
| 5.2 — Optimisation Bas Niveau | 30 min | 30 min | 60 min |
| **Total Module 05** | **60 min** | **65 min** | **~2h** |

---

## Une note éthique

Les outils présentés dans ce module — scanners de ports, tunneling réseau — sont des outils légitimes d'administration système et de sécurité défensive.

Leur utilisation sur des systèmes que vous ne possédez pas ou pour lesquels vous n'avez pas d'autorisation explicite est illégale dans la plupart des juridictions.

Utilisez ces connaissances de façon responsable.

---

<div align="center">

[⬅️ Retour au Module 04](../module-04-architecte/README.md) · [👉 Commencer le Chapitre 5.1](./01-reseau-securite.md)

</div>
