# Introduction

## Avant d'écrire la première ligne de code

Cette introduction est courte. Volontairement.

Elle a un seul objectif : vous mettre en condition pour que la formation soit la plus efficace possible. Pas de théorie ici — juste les informations pratiques dont vous avez besoin avant de commencer.

Lisez-la en entier. Ça prend dix minutes et ça vous en économisera beaucoup plus.

---

## Ce dont vous avez besoin

### Les prérequis réels

Pas besoin de connaître Go. C'est le but de la formation.

En revanche, vous serez plus à l'aise si vous avez déjà :

- Écrit des fonctions dans **n'importe quel langage** (Python, JavaScript, PHP, Java, peu importe)
- Utilisé un **terminal** au moins une fois (ouvrir un dossier, lancer une commande)
- Compris vaguement ce qu'est une **variable** et une **boucle**

C'est tout. Vraiment.

> 💡 **Si vous venez du Python ou du JavaScript :** vous allez parfois trouver Go "verbeux" ou "rigide". C'est normal. Résistez à l'envie de chercher un raccourci — la rigidité de Go est une fonctionnalité, pas un bug. On expliquera pourquoi en chemin.

> 💡 **Si vous venez du Java ou du C++ :** vous allez parfois trouver Go "trop simple" ou "incomplet". C'est normal aussi. L'absence de certaines fonctionnalités est elle aussi délibérée. On y reviendra.

---

### Installer Go

Rendez-vous sur le site officiel : **[go.dev/dl](https://go.dev/dl/)**

Téléchargez la version stable pour votre système d'exploitation et suivez les instructions d'installation. En 2024, Go s'installe en moins de cinq minutes sur n'importe quelle machine.

Une fois installé, vérifiez que tout fonctionne en ouvrant un terminal et en tapant :

```bash
go version
```

Vous devriez voir quelque chose comme :

```
go version go1.23.0 linux/amd64
```

Si c'est le cas, vous êtes prêt.

---

### L'éditeur de code

Go fonctionne avec n'importe quel éditeur. Mais deux sont particulièrement bien supportés :

**VS Code** avec l'extension officielle Go
→ [code.visualstudio.com](https://code.visualstudio.com/) + chercher "Go" dans les extensions

**GoLand** (JetBrains) — payant, mais très complet
→ [jetbrains.com/go](https://www.jetbrains.com/go/)

Pour cette formation, **VS Code suffit largement**. L'extension officielle Go gère l'autocomplétion, le formatage automatique, et la détection d'erreurs à la volée.

---

## Comprendre la structure du dépôt

Voici comment ce dépôt est organisé et comment naviguer dedans efficacement.

```
formation-go/
│
├── README.md           ← La vitrine. Vous l'avez déjà lu.
├── PREFACE.md          ← Le "pourquoi". Vous venez de le lire.
├── INTRODUCTION.md     ← Vous êtes ici.
│
├── module-01-eveil/    ← Commencez ici si vous êtes nouveau
├── module-02-forge/
├── module-03-super-pouvoir/
├── module-04-architecte/
├── module-05-ombre/    ← Module bonus, indépendant
│
└── projets/
    ├── gowatch/        ← Projet CLI fil rouge
    └── gohub/          ← Projet API fil rouge
```

Chaque dossier de module contient :
- Un **`README.md`** qui présente le module, ses objectifs, et ce que vous allez construire
- Un fichier **`.md` par chapitre**, numéroté dans l'ordre à suivre
- Des **exemples de code** intégrés directement dans les chapitres

Les dossiers `projets/` contiennent le code source complet des deux projets fil rouge, organisé par étape correspondant à chaque module.

---

## Quel chemin prendre ?

La formation est conçue pour être suivie dans l'ordre — mais pas forcément en entier.

Voici le parcours recommandé selon votre profil :

### 🔧 Développeur Backend / Fullstack
**Objectif :** construire des APIs Go robustes et les déployer.
```
Module 01 → Module 02 → Module 03 → Module 04
```
Le Module 05 est optionnel pour vous. Revenez-y si vous êtes curieux des internals.

---

### ☁️ Ingénieur DevOps / Cloud
**Objectif :** écrire des outils CLI solides et maîtriser le déploiement Cloud Native.
```
Module 01 → Module 02 → Module 03 → Module 04 → Module 05
```
Accordez une attention particulière aux chapitres 4.3 (Déploiement) et 5.1 (Réseau).

---

### 🐍 Développeur Python / JavaScript
**Objectif :** comprendre la mémoire, la concurrence, et le typage statique sans souffrir.
```
Module 01 → Module 02 → Module 03 → Module 04
```
Soyez particulièrement attentif au chapitre 2.2 (gestion d'erreurs) et 3.1 (goroutines).
Ces deux points sont les plus déroutants en venant de Python ou JS — et les plus libérateurs une fois compris.

---

### 🔬 Curieux Low-Level / Sécurité
**Objectif :** aller près du métal, comprendre le réseau, écrire des outils système.
```
Module 01 → Module 02 → Module 03 → Module 05
```
Le Module 04 est recommandé mais pas obligatoire si votre objectif est purement système.

---

### 🌱 Débutant complet
**Objectif :** apprendre à programmer sérieusement avec un premier langage solide.
```
Module 01 → Module 02 → Module 03 → Module 04
```
Prenez votre temps sur le Module 02. Ne passez pas au suivant tant que les exemples de code ne vous semblent pas naturels. Il n'y a pas de chrono.

---

## La structure de chaque chapitre

Chaque chapitre suit toujours le même schéma. Une fois que vous l'avez repéré, la lecture devient plus fluide :

| Section | Ce que vous y trouvez |
|---------|----------------------|
| **Le problème** | La situation concrète que ce chapitre résout |
| **L'intuition** | L'analogie ou l'image mentale avant le code |
| **La solution Go** | Le code commenté, expliqué ligne par ligne |
| **Ce qu'il faut retenir** | L'essentiel en 3 points maximum |
| **Pour aller plus loin** | Des ressources optionnelles pour creuser |

> ⚠️ **Conseil important :** Ne vous contentez pas de lire le code. Tapez-le. Modifiez-le. Cassez-le volontairement et observez ce que Go vous dit. Le compilateur de Go est un des plus pédagogiques qui soit — ses messages d'erreur sont clairs et précis. Apprenez à les lire comme des alliés, pas comme des ennemis.

---

## Les deux projets fil rouge

Tout au long de la formation, vous allez construire deux outils réels :

### 🖥️ `gowatch` — Outil CLI de monitoring système

Un programme en ligne de commande qui surveille votre machine : CPU, RAM, disque, réseau.

Il commence simple au Module 01 (un binaire qui affiche des infos) et devient, au Module 05, un outil concurrent, sécurisé, distribuable en un seul fichier.

```bash
# Ce que vous saurez faire à la fin
$ gowatch --cpu --ram --interval 2s
```

### 🌐 `gohub` — Dashboard de monitoring via API REST

Une API HTTP qui expose les métriques collectées par `gowatch`, avec un historique en mémoire, des endpoints JSON, et une image Docker de moins de 10 Mo.

```bash
# Ce que vous saurez faire à la fin
$ curl http://localhost:8080/metrics
{"cpu": 23.4, "ram": 67.1, "uptime": "4h32m"}
```

---

## Une convention à connaître

Dans tous les chapitres, vous trouverez des blocs signalés comme ceci :

> 💡 **Astuce** — Une bonne pratique ou un raccourci utile.

> ⚠️ **Attention** — Un piège classique ou une erreur fréquente à éviter.

> 🔍 **Zoom profil** — Une précision spécifique à un type d'apprenant.

> 🛠️ **Projet fil rouge** — Le moment où on applique le chapitre au projet concret.

---

## Vous êtes prêt

Go est installé. Vous savez où vous allez. Vous connaissez votre point d'entrée.

Il reste une chose à faire : commencer.

---

<div align="center">

[👉 Commencer le Module 01 — L'Éveil](./module-01-eveil/README.md)

</div>
