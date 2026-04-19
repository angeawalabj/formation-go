# 🐹 Apprendre Go : Mais Pourquoi ? Et Pour Quoi Faire ?

> *"La simplicité est la sophistication suprême."* — Rob Pike, co-créateur de Go

---

## 📖 À propos de cette formation

Ce dépôt est le contenu complet d'une formation en français sur le langage Go, conçue pour **tous les profils** — du développeur backend frustré par la complexité, à l'ingénieur DevOps qui veut des outils qui ne s'effondrent pas, en passant par le curieux venu du Python ou du JavaScript.

Elle ne commence pas par "voici comment déclarer une variable".

Elle commence par **une vraie question** : pourquoi votre serveur sature à 1 000 connexions avec Python, et comment Go règle ce problème en quelques lignes ?

La théorie arrive **après** la douleur. Toujours.

---

## 🎯 À qui s'adresse cette formation ?

| Profil | Ce que tu vas gagner |
|--------|----------------------|
| 🔧 **Développeur Backend / Fullstack** | Construire des APIs robustes, rapides, légères à déployer |
| ☁️ **Ingénieur DevOps / Cloud** | Écrire des outils CLI et des pipelines d'infrastructure solides |
| 🐍 **Développeur Python / JS** | Comprendre la mémoire, la concurrence, sans souffrir comme en C++ |
| 🔬 **Curieux du Low-Level** | Aller près du métal sans quitter le confort d'un langage moderne |

---

## 🗺️ Plan de la formation

La formation est organisée en **5 modules progressifs**, avec **2 projets fil rouge** qui grandissent tout au long du parcours.

### Module 01 — L'Éveil *(Pourquoi Go existe)*
> Créer le choc. Comprendre le positionnement stratégique de Go avant d'écrire une seule ligne.

- Chapitre 1.1 — Le syndrome de la complexité
- Chapitre 1.2 — La "Zen Attitude" du Gopher

### Module 02 — La Forge *(Les fondamentaux sans s'ennuyer)*
> La syntaxe expliquée par les problèmes qu'elle résout, pas par sa définition.

- Chapitre 2.1 — Les briques fondamentales
- Chapitre 2.2 — La logique sans fioritures
- Chapitre 2.3 — Composition vs Héritage

### Module 03 — Le Super-Pouvoir *(La concurrence massive)*
> Le moment où tout bascule. Goroutines, Channels, et la philosophie qui change tout.

- Chapitre 3.1 — Goroutines : le chaos ordonné
- Chapitre 3.2 — Channels : l'art de communiquer
- Chapitre 3.3 — Patterns de concurrence avancés

### Module 04 — L'Architecte *(Systèmes sérieux en production)*
> Du bac à sable à la production réelle. Tests, déploiement, observabilité.

- Chapitre 4.1 — Cloud Native et Microservices
- Chapitre 4.2 — Tooling et Qualité Production
- Chapitre 4.3 — Déploiement et DevOps

### Module 05 — L'Homme de l'Ombre *(Outils système avancés)* ⭐ Bonus
> Pour ceux qui veulent aller plus loin : réseau, sécurité, bas niveau.

- Chapitre 5.1 — Réseau et Sécurité Système
- Chapitre 5.2 — Optimisation Bas Niveau

---

## 🛠️ Les deux projets fil rouge

Plutôt que des mini-exercices isolés, chaque module contribue à construire **deux vrais outils** :

### 🖥️ Projet A — `gowatch` : Un outil CLI de monitoring système
Un outil en ligne de commande qui surveille votre machine en temps réel.
Du simple binaire du Module 01 jusqu'à l'outil réseau sécurisé du Module 05.

### 🌐 Projet B — `gohub` : Un dashboard de monitoring via API REST
Une API qui expose les métriques collectées, dockerisée, testée, prête à déployer.
Introduite au Module 04, elle s'appuie sur tout ce qui a été construit avant.

---

## 📐 Philosophie pédagogique

Cette formation adopte **trois principes** qui ne bougent pas :

**1. Le Problème d'abord**
Chaque concept est introduit par la frustration qu'il résout. On ne définit pas `goroutine` — on montre d'abord le serveur qui rame, puis on introduit la solution.

**2. L'Épurisme Radical**
Go est né d'un refus de la complexité inutile. Cette formation aussi. Chaque chapitre défend une idée centrale, et une seule.

**3. Le Code qui s'explique tout seul**
Tous les exemples de code sont commentés ligne par ligne. Pas de magie noire. Pas de "faites-moi confiance". Tout est visible, tout est explicable.

---

## 🗂️ Structure du dépôt

```
formation-go/
├── README.md               ← Vous êtes ici
├── PREFACE.md              ← Pourquoi cette formation existe
├── INTRODUCTION.md         ← Comment utiliser ce contenu efficacement
├── module-01-eveil/
│   ├── README.md           ← Vue d'ensemble du module
│   ├── 01-syndrome-complexite.md
│   └── 02-zen-attitude-gopher.md
├── module-02-forge/
│   ├── README.md
│   ├── 01-briques-fondamentales.md
│   ├── 02-logique-sans-fioritures.md
│   └── 03-composition-vs-heritage.md
├── module-03-super-pouvoir/
│   ├── README.md
│   ├── 01-goroutines.md
│   ├── 02-channels.md
│   └── 03-patterns-concurrence.md
├── module-04-architecte/
│   ├── README.md
│   ├── 01-cloud-native-microservices.md
│   ├── 02-tooling-qualite.md
│   └── 03-deploiement-devops.md
├── module-05-ombre/
│   ├── README.md
│   ├── 01-reseau-securite.md
│   └── 02-optimisation-bas-niveau.md
└── projets/
    ├── gowatch/            ← Code source du projet CLI
    └── gohub/              ← Code source du projet API
```

---

## 🚀 Comment commencer ?

**Si tu es complètement nouveau sur Go :**
→ Commence par la [Préface](./PREFACE.md), puis l'[Introduction](./INTRODUCTION.md), puis le [Module 01](./module-01-eveil/).

**Si tu connais déjà les bases :**
→ Lis l'[Introduction](./INTRODUCTION.md) pour trouver ton point d'entrée selon ton profil.

**Si tu veux juste les projets :**
→ Va directement dans le dossier [`/projets`](./projets/).

---

## 📋 Prérequis

- Aucune connaissance de Go requise
- Une expérience minimale dans **n'importe quel** langage de programmation
- Go installé sur votre machine → [Installation officielle](https://go.dev/dl/)
- Un terminal et la curiosité de comprendre comment les choses fonctionnent vraiment

---

## 📄 Licence

Ce contenu est partagé sous licence **Creative Commons CC BY-NC-SA 4.0**.
Vous pouvez le lire, le partager, l'adapter — pas le revendre tel quel.

---

<div align="center">

**Prêt à changer de paradigme ?**

[👉 Commencer par la Préface](./PREFACE.md)

</div>
