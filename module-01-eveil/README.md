# Module 01 — L'Éveil

> *"Les langages complexes sont nés de besoins simples mal résolus."*

---

## Ce que ce module va changer

Avant ce module, Go est peut-être pour vous "un langage de plus à apprendre".

Après ce module, vous comprendrez pourquoi des entreprises comme **Google, Docker, Uber, Cloudflare et Kubernetes** ont fait le choix de Go — et pourquoi ce choix n'est pas anodin.

Ce module ne vous apprend pas à coder en Go.

Il vous apprend à **penser en Gopher**.

C'est la différence entre savoir conduire une voiture et comprendre pourquoi elle a été conçue ainsi. La conduite viendra plus vite, et mieux, si vous comprenez d'abord la machine.

---

## Les objectifs de ce module

À la fin du Module 01, vous serez capable de :

- ✅ Expliquer pourquoi Go a été créé et quels problèmes concrets il résout
- ✅ Positionner Go par rapport aux autres langages (Python, Java, Rust, C++)
- ✅ Identifier les situations où Go est le bon choix — et celles où il ne l'est pas
- ✅ Installer et configurer votre environnement de travail Go
- ✅ Compiler et exécuter votre premier programme Go
- ✅ Comprendre ce que fait la toolchain Go (build, run, fmt, vet)

---

## Ce que vous allez construire

> 🛠️ **Projet fil rouge — Étape 1 : Initialisation de `gowatch`**

À la fin de ce module, vous aurez compilé la première version de `gowatch` — votre outil CLI de monitoring système.

Pour l'instant, il ne fait qu'une chose : afficher des informations sur votre machine. Mais il sera déjà :

- Compilé en un **binaire unique** sans dépendances
- Exécutable sur n'importe quelle machine du même système d'exploitation
- Structuré pour accueillir toutes les fonctionnalités des modules suivants

```bash
$ ./gowatch
Système  : linux/amd64
Go       : go1.23.0
Hostname : mon-serveur
```

Petit. Propre. Fonctionnel. C'est très Go.

---

## Les chapitres de ce module

### [Chapitre 1.1 — Le syndrome de la complexité](./01-syndrome-complexite.md)
On commence par le problème. Pourquoi les langages traditionnels peinent face aux exigences du Cloud moderne ? Qu'est-ce que Rob Pike et Ken Thompson ont voulu résoudre en créant Go ? Et qu'est-ce que ça change pour vous, concrètement ?

**Concepts abordés :** Historique de Go, comparaison des langages, les 3 piliers fondateurs.

---

### [Chapitre 1.2 — La "Zen Attitude" du Gopher](./02-zen-attitude-gopher.md)
On installe, on configure, on écrit le premier programme. Mais surtout, on comprend ce que chaque ligne de ce premier programme révèle sur la philosophie du langage. Le "Hello World" de Go n'est pas anodin.

**Concepts abordés :** Installation, toolchain, premier programme, décision d'architecture.

---

## L'état d'esprit à adopter pour ce module

Ce module demande une chose difficile pour beaucoup de développeurs expérimentés : **mettre de côté ce qu'on sait**.

Si vous venez de Java, oubliez temporairement les classes et l'héritage.
Si vous venez de Python, oubliez temporairement la flexibilité du typage dynamique.
Si vous venez du JavaScript, oubliez temporairement les promesses et l'asynchrone.

Go a des réponses différentes à ces problèmes. Pas forcément meilleures dans l'absolu — mais cohérentes, réfléchies, et souvent surprenantes d'efficacité.

Donnez-lui le bénéfice du doute jusqu'au Module 03. Après ça, vous serez convaincu par vous-même.

---

## Durée estimée

| Chapitre | Lecture | Pratique | Total |
|----------|---------|----------|-------|
| 1.1 — Le syndrome de la complexité | 20 min | 10 min | 30 min |
| 1.2 — La "Zen Attitude" du Gopher | 20 min | 30 min | 50 min |
| **Total Module 01** | **40 min** | **40 min** | **~1h20** |

---

## Ressources complémentaires

Ces ressources sont **optionnelles**. Elles enrichissent le module mais ne sont pas nécessaires pour continuer.

- 🎥 [Go at Google: Language Design in the Service of Software Engineering](https://talks.golang.org/2012/splash.article) — Rob Pike (en anglais)
- 📄 [Effective Go](https://go.dev/doc/effective_go) — La référence idiomatique officielle (en anglais)
- 🎙️ [La genèse de Go expliquée par ses créateurs](https://go.dev/blog/10years) — Le blog officiel Go (en anglais)

---

<div align="center">

[👉 Commencer le Chapitre 1.1 — Le syndrome de la complexité](./01-syndrome-complexite.md)

</div>
