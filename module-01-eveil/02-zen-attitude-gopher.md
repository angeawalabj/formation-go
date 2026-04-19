# Chapitre 1.2 — La "Zen Attitude" du Gopher

> *"Go n'essaie pas d'être élégant. Il essaie d'être utile."*

---

## Le problème

Vous avez compris pourquoi Go existe. Maintenant, une question pratique se pose :

**Comment commence-t-on ?**

Et surtout — est-ce que "commencer" va ressembler à l'installation de Java, où l'on passe une demi-heure à configurer le `JAVA_HOME`, choisir entre JDK et JRE, et se battre avec les variables d'environnement ?

La réponse est non. Et cette réponse dit déjà beaucoup sur Go.

---

## L'intuition

Go applique à son propre écosystème la même philosophie qu'il applique au code : **la simplicité d'abord**.

Installer Go, c'est télécharger un archive, l'extraire, et ajouter un dossier au PATH. C'est tout. Pas de registre système à modifier. Pas de machine virtuelle à configurer. Pas de choix de version à faire entre dix distributions différentes.

Et une fois installé, Go vous donne une **toolchain complète** — un ensemble d'outils intégrés qui couvrent la compilation, le formatage, les tests, la gestion des dépendances, et plus encore. Pas besoin de chercher des outils tiers pour faire du Go professionnel. Tout est déjà là.

Pensez-y comme un couteau suisse livré avec le langage. Compact, complet, immédiatement opérationnel.

---

## La solution Go

### Étape 1 — Installation

Rendez-vous sur **[go.dev/dl](https://go.dev/dl/)** et téléchargez la version stable pour votre système.

**Sur Linux / macOS :**
```bash
# Télécharger et extraire (remplacez la version si nécessaire)
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# Ajouter Go au PATH (dans ~/.bashrc ou ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin

# Recharger le terminal
source ~/.bashrc
```

**Sur Windows :**
Téléchargez le fichier `.msi` et suivez l'assistant d'installation. Go sera automatiquement ajouté au PATH.

**Vérification :**
```bash
go version
# go version go1.23.0 linux/amd64
```

Si vous voyez cette ligne, Go est installé. C'est vraiment aussi simple que ça.

---

### Étape 2 — La toolchain : vos outils du quotidien

Go est livré avec une série de commandes que vous utiliserez tous les jours. Voici les essentielles :

```bash
go run fichier.go      # Compile ET exécute en une commande (pour le développement)
go build               # Compile et produit un binaire exécutable
go fmt ./...           # Formate automatiquement tout le code selon les standards Go
go vet ./...           # Analyse le code à la recherche d'erreurs logiques
go test ./...          # Lance tous les tests du projet
go mod init nom        # Initialise un nouveau module Go (votre projet)
go get package         # Télécharge et installe une dépendance externe
```

> 💡 **Astuce** — `./...` est une notation Go qui signifie "ce dossier et tous ses sous-dossiers". Vous la verrez partout. Retenez-la.

> ⚠️ **Attention** — En Go, il n'y a pas de `go compile` ou `go execute`. On dit `go build` et `go run`. C'est une convention mineure, mais mieux vaut le savoir dès le départ pour ne pas chercher des commandes qui n'existent pas.

---

### Étape 3 — Votre premier projet Go

En Go, tout projet commence par l'initialisation d'un **module**. Un module, c'est simplement la déclaration que "ce dossier est un projet Go, et voici son nom".

```bash
# Créer un dossier pour le projet
mkdir mon-premier-go
cd mon-premier-go

# Initialiser le module
go mod init mon-premier-go
```

Cette commande crée un fichier `go.mod` :

```
module mon-premier-go

go 1.23.0
```

Deux lignes. Le nom du module, la version de Go. C'est tout ce dont Go a besoin pour savoir que ce dossier est un projet.

> 🔍 **Zoom profil — Développeurs Node.js**
> `go.mod` est l'équivalent Go de `package.json`. Mais il est généré automatiquement, et vous n'avez presque jamais besoin de l'éditer à la main. Go s'en occupe.

---

### Étape 4 — Le premier programme Go, analysé ligne par ligne

Créez un fichier `main.go` dans votre dossier :

```go
package main

import "fmt"

func main() {
    fmt.Println("Bonjour, monde.")
}
```

Lancez-le :

```bash
go run main.go
# Bonjour, monde.
```

Six lignes. Mais chacune révèle quelque chose d'important sur Go.

---

#### `package main`

Tout fichier Go appartient à un **package** — un regroupement logique de code.

`main` est un package spécial : c'est le point d'entrée de tout programme exécutable. Si votre fichier ne commence pas par `package main`, Go sait que c'est une bibliothèque, pas un programme.

Cette distinction est explicite, visible dès la première ligne. C'est volontaire : **en Go, rien n'est caché**.

> 🔍 **Zoom profil — Développeurs Java**
> En Java, la classe qui contient `public static void main(String[] args)` est le point d'entrée. Go fait la même chose, mais au niveau du package plutôt que de la classe. Plus simple, plus visible.

---

#### `import "fmt"`

Go importe explicitement chaque dépendance utilisée.

`fmt` est un package de la bibliothèque standard de Go — il gère le formatage et l'affichage. Son nom vient de *format*.

Ce qui est remarquable : **si vous importez un package et ne l'utilisez pas, Go refuse de compiler**. C'est une erreur, pas un avertissement.

```go
import "fmt"
import "os"  // ← Si vous n'utilisez pas 'os', le code ne compile pas
```

C'est frustrant les premières fois. Mais c'est délibéré : Go veut que chaque ligne de votre code soit là pour une raison. Pas de code mort. Pas d'imports "au cas où".

> ⚠️ **Attention** — C'est probablement la première "friction" que vous allez rencontrer en apprenant Go. Résistez à l'envie de la contourner. Elle vous sauvera de nombreux bugs à long terme.

---

#### `func main()`

La fonction `main` est le point d'entrée du programme. Quand vous lancez votre binaire Go, c'est cette fonction qui s'exécute en premier.

Notez la syntaxe : `func` pour déclarer une fonction, le nom, les parenthèses, les accolades. Pas de `public`, pas de `static`, pas de type de retour obligatoire. Juste ce qui est nécessaire.

---

#### `fmt.Println("Bonjour, monde.")`

`fmt.Println` appelle la fonction `Println` du package `fmt`. Elle affiche le texte passé en argument et ajoute un retour à la ligne.

La notation `package.Fonction` est universelle en Go. C'est toujours ainsi qu'on appelle une fonction d'un package externe. Jamais d'ambiguïté sur l'origine d'une fonction.

---

### Étape 5 — Compiler un vrai binaire

`go run` est pratique pour le développement. Mais pour distribuer votre programme, on utilise `go build` :

```bash
go build -o mon-programme main.go

# Exécuter le binaire produit
./mon-programme
# Bonjour, monde.
```

Le fichier `mon-programme` produit est un **binaire natif autonome**. Copiez-le sur n'importe quelle machine Linux du même type de processeur — il fonctionnera sans rien installer. Pas de Go. Pas de runtime. Pas de dépendances.

C'est une des caractéristiques les plus précieuses de Go dans un contexte de déploiement moderne.

---

### Étape 6 — Le formatage automatique

Tapez cette commande dans votre projet :

```bash
go fmt ./...
```

Go reformate automatiquement votre code selon les conventions officielles du langage. Indentation, espaces, sauts de ligne — tout est standardisé.

Testez-le : mettez votre code dans n'importe quel état (mauvaise indentation, espaces incohérents), et lancez `go fmt`. Il le remet en ordre.

> 💡 **Astuce** — Configurez votre éditeur pour lancer `go fmt` automatiquement à chaque sauvegarde. Quasiment tous les plugins Go pour VS Code et GoLand le font par défaut. Une fois habitué, vous ne pourrez plus vous en passer.

---

## 🛠️ Projet fil rouge — Initialisation de `gowatch`

Maintenant qu'on maîtrise les bases de la toolchain, on initialise le projet fil rouge.

Voici ce qu'on va construire dans ce chapitre : la première version de `gowatch`, un outil CLI qui affiche des informations sur le système.

**Créez la structure suivante :**

```bash
mkdir gowatch
cd gowatch
go mod init github.com/votre-pseudo/gowatch
```

> 💡 **Astuce** — Par convention, les modules Go sont nommés avec l'URL de leur dépôt GitHub, même si le projet n'est pas encore en ligne. C'est la bonne habitude à prendre dès le début.

**Créez le fichier `main.go` :**

```go
package main

import (
    "fmt"
    "runtime"
)

func main() {
    // runtime est un package de la bibliothèque standard
    // Il donne accès aux informations sur l'environnement d'exécution Go
    fmt.Println("=== gowatch v0.1 ===")
    fmt.Println("")

    // GOOS : le système d'exploitation (linux, darwin, windows)
    // GOARCH : l'architecture du processeur (amd64, arm64...)
    fmt.Printf("Système  : %s/%s\n", runtime.GOOS, runtime.GOARCH)

    // Version de Go utilisée pour compiler ce binaire
    fmt.Printf("Go       : %s\n", runtime.Version())

    // Nombre de CPU logiques disponibles sur la machine
    fmt.Printf("CPU(s)   : %d\n", runtime.NumCPU())

    // Nombre de goroutines actuellement en cours d'exécution
    // (pour l'instant, il n'y en a qu'une : main elle-même)
    fmt.Printf("Goroutines actives : %d\n", runtime.NumGoroutine())
}
```

**Compilez et exécutez :**

```bash
go run main.go
```

**Résultat attendu :**

```
=== gowatch v0.1 ===

Système  : linux/amd64
Go       : go1.23.0
CPU(s)   : 8
Goroutines actives : 1
```

**Maintenant, compilez un vrai binaire :**

```bash
go build -o gowatch main.go
./gowatch
```

Vous venez de créer votre premier outil CLI en Go. Il est compilé, autonome, et prêt à être distribué.

Dans les modules suivants, `gowatch` va s'enrichir : collecte de métriques CPU et RAM, affichage en temps réel, concurrent, puis exposé via une API REST. Mais sa fondation, c'est ce que vous venez d'écrire.

---

## Quand choisir Go ? — La décision d'architecture

Avant de clore ce module, voici un guide de décision rapide que vous pourrez réutiliser dans vos projets futurs :

```
Votre besoin principal est...

Performance brute + concurrence massive ?
    → Go est un excellent choix

Binaire unique, déploiement sans dépendances ?
    → Go est fait pour ça

Script rapide, automatisation ponctuelle ?
    → Python est plus rapide à écrire pour ça

Machine learning, data science ?
    → Python et son écosystème (NumPy, PyTorch) sont imbattables

Interface graphique desktop ?
    → Go peut le faire, mais d'autres écosystèmes sont plus matures

Sécurité mémoire absolue, proche du noyau ?
    → Rust est plus adapté

API web standard, équipe habituée à Java ?
    → Go est une excellente alternative, mais le contexte d'équipe compte
```

> 💡 **Le vrai critère** — La meilleure décision technique n'est pas toujours celle qui choisit le langage "objectivement meilleur". C'est celle qui prend en compte la performance, le déploiement, la maintenabilité **et** les compétences de l'équipe. Go excelle quand vous avez besoin de performance, de simplicité de déploiement, et d'une concurrence robuste. Si ces trois critères sont présents, Go mérite sérieusement d'être considéré.

---

## Ce qu'il faut retenir

1. **L'installation de Go est délibérément simple** — c'est le reflet de la philosophie du langage. Ce qui est vrai pour l'installation est vrai pour tout le reste.

2. **La toolchain intégrée est complète** — `go run`, `go build`, `go fmt`, `go test` couvrent 95% de vos besoins sans rien installer d'autre.

3. **Chaque ligne de votre premier programme révèle quelque chose** — les imports explicites, le package main, l'absence de boilerplate. Tout est intentionnel.

4. **Un binaire Go est autonome** — pas de runtime, pas de dépendances. Compilez une fois, déployez partout.

---

## Pour aller plus loin

- 📄 [A Tour of Go](https://go.dev/tour/) — Le tutoriel interactif officiel, directement dans le navigateur
- 📄 [How to Write Go Code](https://go.dev/doc/code) — La documentation officielle sur la structure d'un projet Go
- 🔧 [Extension Go pour VS Code](https://marketplace.visualstudio.com/items?itemName=golang.go) — L'extension officielle avec autocomplétion et formatage automatique

---

<div align="center">

[⬅️ Chapitre 1.1 — Le syndrome de la complexité](./01-syndrome-complexite.md) · [👉 Module 02 — La Forge](../module-02-forge/README.md)

</div>
