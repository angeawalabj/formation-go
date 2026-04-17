# Module 03 — Le Super-Pouvoir

> *"Do not communicate by sharing memory; instead, share memory by communicating."*
> — Rob Pike

---

## Ce que ce module va changer

Les Modules 01 et 02 vous ont donné la philosophie et les outils de base.

Ce module vous donne ce que personne d'autre ne donne aussi simplement : **la capacité de penser et d'écrire des programmes concurrents**.

La concurrence — exécuter plusieurs choses en même temps — est souvent présentée comme un sujet réservé aux experts. Quelque chose qu'on aborde "plus tard", "quand on est prêt", avec des avertissements sur les deadlocks, les race conditions, et la complexité des threads.

Go renverse cette logique. La concurrence n'est pas une fonctionnalité avancée dans Go — c'est une **fonctionnalité de base**, aussi naturelle qu'une boucle ou une condition.

Ce module est celui où la plupart des développeurs "tombent amoureux" de Go. Préparez-vous.

---

## Les objectifs de ce module

À la fin du Module 03, vous serez capable de :

- ✅ Lancer des tâches concurrentes avec le mot-clé `go`
- ✅ Comprendre la différence fondamentale entre un thread OS et une goroutine
- ✅ Faire communiquer des goroutines de façon sûre via des channels
- ✅ Orchestrer des flux de données avec `select`
- ✅ Synchroniser des goroutines avec `WaitGroup` et `Mutex`
- ✅ Propager des annulations et des timeouts avec le `Context`
- ✅ Identifier et éviter les pièges classiques : goroutine leak, race condition, deadlock

---

## Ce que vous allez construire

> 🛠️ **Projet fil rouge — Étape 3 : `gowatch` devient concurrent**

À la fin de ce module, `gowatch` collectera ses métriques **en parallèle** depuis plusieurs sources simultanément, les agrégera via des channels, et gérera les timeouts proprement.

```bash
$ ./gowatch --watch
[15:42:01] CPU=23.4%  RAM=67.1%  Goroutines=4  Latence=2ms
[15:42:03] CPU=31.2%  RAM=67.3%  Goroutines=4  Latence=1ms
[15:42:05] CPU=18.9%  RAM=67.0%  Goroutines=4  Latence=3ms
^C Arrêt propre. 3 snapshots collectés.
```

La collecte de chaque métrique se fera dans sa propre goroutine. Les résultats transiteront par des channels. Un timeout via Context arrêtera proprement le tout si une source tarde à répondre.

---

## Les chapitres de ce module

### [Chapitre 3.1 — Goroutines : le chaos ordonné](./01-goroutines.md)
Le mot-clé `go`, le scheduler de Go, la différence entre thread OS et goroutine. Et les pièges à éviter dès le départ.

**Concepts abordés :** `go`, goroutine leak, race condition, `go race detector`.

---

### [Chapitre 3.2 — Channels : l'art de communiquer](./02-channels.md)
Comment faire passer des données entre goroutines de façon sûre. Channels buffered et unbuffered, `select`, patterns de fermeture.

**Concepts abordés :** `chan`, `<-`, channels buffered/unbuffered, `select`, `close`, `range` sur channel.

---

### [Chapitre 3.3 — Patterns de concurrence avancés](./03-patterns-concurrence.md)
Les outils de synchronisation quand les channels ne suffisent pas. Et le Context API pour contrôler la durée de vie des goroutines.

**Concepts abordés :** `sync.WaitGroup`, `sync.Mutex`, `sync.Map`, `context.Context`, timeout, annulation.

---

## Durée estimée

| Chapitre | Lecture | Pratique | Total |
|----------|---------|----------|-------|
| 3.1 — Goroutines | 25 min | 25 min | 50 min |
| 3.2 — Channels | 30 min | 30 min | 60 min |
| 3.3 — Patterns avancés | 30 min | 35 min | 65 min |
| **Total Module 03** | **85 min** | **90 min** | **~3h** |

---

## L'état d'esprit à adopter

La concurrence demande un changement de perspective.

Jusqu'ici, vous avez pensé votre code comme une séquence d'instructions : "fais A, puis B, puis C". La concurrence, c'est : "commence A, B et C en même temps, et récupère les résultats quand ils sont prêts."

Ce n'est pas plus compliqué — c'est différent. Donnez-vous le temps de laisser ce paradigme s'installer. Les exemples de code de ce module sont progressifs : chaque concept est introduit seul, avant d'être combiné avec le suivant.

Une fois que vous aurez écrit votre premier programme concurrent qui fonctionne, vous ne voudrez plus jamais revenir en arrière.

---

<div align="center">

[⬅️ Retour au Module 02](../module-02-forge/README.md) · [👉 Commencer le Chapitre 3.1](./01-goroutines.md)

</div>
