# Chapitre 2.2 — La logique sans fioritures

> *"Les exceptions sont une façon élégante de cacher les erreurs. Go préfère les montrer."*

---

## Le problème

Ouvrez n'importe quelle base de code Java ou Python de taille moyenne.

Vous trouverez des `try/catch` imbriqués les uns dans les autres. Des exceptions lancées dans une fonction, rattrapées trois niveaux plus haut, re-lancées sous un autre nom. Des blocs `finally` qui s'exécutent "quoi qu'il arrive" — sauf quand ils ne s'exécutent pas, pour des raisons obscures.

Vous trouverez aussi des boucles `for`, des boucles `while`, des boucles `do/while`, des `forEach`, des compréhensions de liste, des générateurs. Chacun avec ses subtilités, ses cas limites, son comportement légèrement différent.

Go a regardé tout ça et a dit : **non**.

Une seule boucle. Pas d'exceptions. Et une gestion d'erreurs si explicite qu'elle en devient presque brutale.

Ce chapitre explique pourquoi ces choix — qui semblent des limitations — sont en réalité des décisions d'ingénierie parmi les plus mûres du langage.

---

## L'intuition

### Sur les boucles

Imaginez un chef de cuisine qui a dans sa cuisine : une casserole, une poêle, un wok, une cocotte, une friteuse, et une marmite à pression.

Il peut tout cuisiner. Mais chaque outil a ses propres règles, ses propres températures, ses propres temps de cuisson. Un apprenti qui arrive dans cette cuisine doit apprendre six outils différents avant de cuisiner quoi que ce soit.

Maintenant imaginez une cuisine avec une seule casserole universelle — assez intelligente pour se comporter comme n'importe lequel de ces outils selon comment vous l'utilisez.

C'est le `for` de Go.

### Sur les erreurs

Dans la plupart des langages, une erreur est une **exception** — quelque chose d'anormal qui "remonte" automatiquement la pile d'appels jusqu'à ce que quelqu'un la rattrape. Ou pas. Et si personne ne la rattrape, le programme plante.

En Go, une erreur est une **valeur** — comme un entier ou une chaîne. Elle ne remonte rien. Elle ne fait rien toute seule. C'est vous qui décidez quoi en faire, explicitement, à chaque niveau.

C'est la différence entre un système d'alarme qui déclenche automatiquement les pompiers (exceptions) et un voyant rouge sur votre tableau de bord que vous devez décider d'ignorer ou de traiter (erreurs Go).

---

## La solution Go

### La boucle `for` — Une seule pour les gouverner toutes

Go n'a qu'un seul mot-clé de boucle : `for`. Mais il s'adapte à tous les cas d'usage.

---

**Forme 1 — La boucle classique avec compteur**
```go
for i := 0; i < 5; i++ {
    fmt.Println(i)
}
// 0, 1, 2, 3, 4
```

C'est le `for` traditionnel : initialisation, condition, incrément. Identique au C, Java, ou JavaScript.

---

**Forme 2 — La boucle `while` (qui n'existe pas en Go)**

Go n'a pas de `while`. Mais un `for` avec seulement une condition se comporte exactement comme un `while` :

```go
compteur := 0
for compteur < 10 {
    compteur++
}
```

> 🔍 **Zoom profil — Développeurs Python**
> L'équivalent de `while True:` en Go s'écrit `for {` — un `for` sans condition. C'est la boucle infinie idiomatique de Go.

---

**Forme 3 — La boucle infinie**
```go
for {
    // S'exécute indéfiniment
    // On en sort avec 'break' ou 'return'
    donnee := lireEntree()
    if donnee == "quitter" {
        break
    }
    traiter(donnee)
}
```

---

**Forme 4 — La boucle `range` sur un Slice**
```go
langages := []string{"Go", "Python", "Rust", "Java"}

for index, langage := range langages {
    fmt.Printf("%d : %s\n", index, langage)
}
// 0 : Go
// 1 : Python
// 2 : Rust
// 3 : Java
```

---

**Forme 5 — La boucle `range` sur une Map**
```go
scores := map[string]int{
    "Alice": 95,
    "Bob":   87,
    "Carol": 92,
}

for nom, score := range scores {
    fmt.Printf("%s a obtenu %d points\n", nom, score)
}
```

---

**Forme 6 — La boucle `range` sur une chaîne de caractères**
```go
mot := "Gopher"

for index, caractere := range mot {
    fmt.Printf("Position %d : %c\n", index, caractere)
}
// Position 0 : G
// Position 1 : o
// Position 2 : p
// Position 3 : h
// Position 4 : e
// Position 5 : r
```

> 💡 **Astuce** — Quand vous itérez sur une chaîne avec `range`, vous obtenez des **runes** (des caractères Unicode complets), pas des bytes. C'est la bonne façon de traiter des chaînes contenant des caractères accentués ou des emojis.

---

**`break` et `continue` — Contrôler le flux**
```go
for i := 0; i < 10; i++ {
    if i == 3 {
        continue  // Sauter cette itération, passer à la suivante
    }
    if i == 7 {
        break     // Sortir complètement de la boucle
    }
    fmt.Println(i)
}
// 0, 1, 2, 4, 5, 6
```

---

### Les conditions — Sobres et efficaces

Go n'a pas de surprises ici — les conditions fonctionnent comme dans la plupart des langages, avec quelques particularités utiles.

```go
age := 25

// Condition simple
if age >= 18 {
    fmt.Println("Majeur")
} else {
    fmt.Println("Mineur")
}
```

> ⚠️ **Attention** — En Go, les parenthèses autour de la condition ne sont **pas** nécessaires (et `gofmt` les supprimera si vous les mettez). Les accolades, en revanche, sont **obligatoires** — même pour un seul statement. Pas de `if condition: instruction` sur une ligne comme en Python.

---

**La condition avec initialisation — un bijou de Go**

Go permet d'initialiser une variable directement dans la condition `if`. Cette variable n'existe que dans le bloc `if/else` :

```go
// La variable 'err' est déclarée ET vérifiée en une seule ligne
if err := faireQuelqueChose(); err != nil {
    fmt.Println("Erreur :", err)
    return
}
// 'err' n'existe plus ici — elle est hors scope
```

C'est élégant et ça évite de polluer le scope avec des variables temporaires.

---

**Le `switch` — Plus puissant qu'il n'y paraît**

```go
jour := "lundi"

switch jour {
case "lundi", "mardi", "mercredi", "jeudi", "vendredi":
    fmt.Println("Jour de semaine")
case "samedi", "dimanche":
    fmt.Println("Week-end")
default:
    fmt.Println("Jour inconnu")
}
```

> 💡 **Astuce** — En Go, le `switch` ne nécessite **pas** de `break` entre les cases. Chaque case s'arrête automatiquement. C'est l'inverse du C ou Java où l'oubli d'un `break` est un bug classique. Si vous voulez le comportement "tomber dans le case suivant", utilisez explicitement `fallthrough`.

---

### La gestion d'erreurs — Le choix le plus controversé de Go

Voici le moment qui divise.

Dans presque tous les langages modernes, les erreurs sont des **exceptions** qui se propagent automatiquement. En Go, les erreurs sont des **valeurs** que vous retournez et traitez manuellement.

Concrètement, une fonction Go qui peut échouer retourne **deux valeurs** : le résultat, et une erreur.

```go
// os.Open retourne un fichier ET une erreur
fichier, err := os.Open("config.json")
if err != nil {
    // Traiter l'erreur ici
    fmt.Println("Impossible d'ouvrir le fichier :", err)
    return
}
// Si on arrive ici, le fichier est ouvert avec succès
defer fichier.Close()  // On y revient juste après
```

Le `if err != nil` est la construction la plus vue en Go. Elle deviendra votre réflexe automatique.

---

**Pourquoi pas de Try/Catch ?**

La réponse de Go est philosophique autant que technique :

Les exceptions créent des **chemins d'exécution invisibles**. Quand une fonction lève une exception, vous ne savez pas, en lisant le code, où elle sera rattrapée — ni si elle le sera. L'exception traverse des couches de code sans laisser de trace dans le flux de lecture.

En Go, **chaque erreur possible est visible dans la signature de la fonction**. Vous savez exactement quelles fonctions peuvent échouer, et vous êtes forcé de décider quoi faire dans chaque cas.

```go
// ✅ En Go : explicite, visible, traçable
resultat, err := calculerQuelqueChose()
if err != nil {
    return fmt.Errorf("calcul impossible : %w", err)
}

// ❌ Équivalent en Java/Python : l'erreur est invisible dans la signature
try {
    resultat = calculerQuelqueChose()  // Peut lever n'importe quelle exception
} catch (Exception e) {
    // Quelqu'un, quelque part, a peut-être documenté ce qui peut arriver ici
}
```

> 🔍 **Zoom profil — Développeurs Java/Python**
> Oui, `if err != nil` répété partout semble verbeux. Et il l'est, objectivement. Mais voici la question à se poser : dans votre code actuel, combien d'exceptions silencieuses attrapez-vous avec un `except Exception: pass` ou un `catch (Exception e) {}` vide ? Combien de bugs en production viennent d'erreurs ignorées ? Go rend l'ignorance d'une erreur visible et délibérée — pas accidentelle.

---

**Créer ses propres erreurs**

```go
import "errors"
import "fmt"

// Erreur simple
err1 := errors.New("quelque chose a mal tourné")

// Erreur avec contexte formaté (la façon moderne)
nom := "config.json"
err2 := fmt.Errorf("impossible de lire le fichier %s : %w", nom, err1)

fmt.Println(err2)
// impossible de lire le fichier config.json : quelque chose a mal tourné
```

Le `%w` dans `fmt.Errorf` est important : il **encapsule** l'erreur originale, permettant de la retrouver plus tard avec `errors.Unwrap` ou `errors.Is`. C'est la façon idiomatique d'ajouter du contexte à une erreur sans perdre l'information originale.

---

**Propager une erreur vers le haut**

```go
func lireConfiguration(chemin string) (string, error) {
    fichier, err := os.Open(chemin)
    if err != nil {
        // On enveloppe l'erreur avec du contexte et on la remonte
        return "", fmt.Errorf("lireConfiguration : %w", err)
    }
    defer fichier.Close()

    // ... lire le contenu ...
    return contenu, nil  // nil signifie "pas d'erreur"
}

func main() {
    config, err := lireConfiguration("app.json")
    if err != nil {
        fmt.Println("Erreur fatale :", err)
        os.Exit(1)  // Quitter le programme avec un code d'erreur
    }
    fmt.Println("Configuration chargée :", config)
}
```

> 💡 **Convention Go** — Quand une fonction réussit, elle retourne `nil` comme erreur. `nil` en Go signifie "absence de valeur" — c'est l'équivalent de `null` en Java ou `None` en Python, mais utilisé ici pour signifier "pas d'erreur".

---

### `defer` — Exécuter du code "à la sortie"

`defer` est une des fonctionnalités les plus élégantes de Go. Il permet de planifier l'exécution d'une fonction **juste avant que la fonction courante se termine** — quelle que soit la façon dont elle se termine (retour normal, retour sur erreur, panic).

```go
func lireFichier(chemin string) error {
    fichier, err := os.Open(chemin)
    if err != nil {
        return err
    }
    defer fichier.Close()  // Sera exécuté quand lireFichier() se terminera

    // ... traiter le fichier ...
    // Pas besoin de se souvenir de fermer le fichier plus bas
    return nil
}
```

Sans `defer`, vous devriez appeler `fichier.Close()` avant chaque `return` — et en oublier un serait une fuite de ressource. Avec `defer`, c'est géré automatiquement, une fois, au bon endroit.

```go
// Autre usage classique : mesurer le temps d'exécution d'une fonction
func maFonctionLente() {
    debut := time.Now()
    defer func() {
        fmt.Printf("Durée : %v\n", time.Since(debut))
    }()

    // ... code long ...
}
```

> ⚠️ **Attention** — Les `defer` s'exécutent dans l'ordre **LIFO** (Last In, First Out) — le dernier `defer` déclaré s'exécute en premier. Si vous avez plusieurs `defer` dans une fonction, gardez cet ordre en tête.

---

### `panic` et `recover` — Le dernier recours

Go a bien un mécanisme pour les erreurs **vraiment** catastrophiques : `panic`.

Un `panic` arrête l'exécution normale du programme et remonte la pile d'appels — un peu comme une exception non rattrapée. La différence : en Go, `panic` est réservé aux situations **véritablement exceptionnelles** (bug dans le code, état impossible), pas aux erreurs métier normales.

```go
func diviser(a, b int) int {
    if b == 0 {
        panic("division par zéro : c'est un bug dans le code appelant")
    }
    return a / b
}
```

Pour rattraper un `panic`, on utilise `recover` — mais uniquement dans une fonction `defer` :

```go
func executerSansPlanter(f func()) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Panic rattrapé :", r)
        }
    }()
    f()
}
```

> ⚠️ **Règle d'or** — Utilisez `panic` uniquement pour les bugs de programmation (un état qui "ne devrait jamais arriver"). Pour toutes les erreurs prévisibles (fichier absent, réseau indisponible, données invalides), utilisez le retour d'erreur normal. Un code Go bien écrit ne `panic` presque jamais.

---

## 🛠️ Projet fil rouge — `gowatch` avec gestion d'erreurs robuste

On refactorise `gowatch` pour séparer la logique de collecte et gérer les erreurs proprement.

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "runtime"
)

// collecterMetriques rassemble les données système
// Elle retourne une map ET une erreur potentielle — idiome Go
func collecterMetriques() (map[string]interface{}, error) {
    // Simuler une erreur possible (ex: lecture système indisponible)
    if runtime.NumCPU() == 0 {
        return nil, fmt.Errorf("collecterMetriques : impossible de lire les infos CPU")
    }

    metriques := map[string]interface{}{
        "systeme":    runtime.GOOS + "/" + runtime.GOARCH,
        "go_version": runtime.Version(),
        "cpu_count":  runtime.NumCPU(),
        "goroutines": runtime.NumGoroutine(),
    }

    return metriques, nil
}

// afficherJSON sérialise et affiche les métriques en JSON
func afficherJSON(metriques map[string]interface{}) error {
    donnees, err := json.MarshalIndent(metriques, "", "  ")
    if err != nil {
        return fmt.Errorf("afficherJSON : %w", err)
    }
    fmt.Println(string(donnees))
    return nil
}

// afficherTexte affiche les métriques en format lisible
func afficherTexte(metriques map[string]interface{}) {
    fmt.Println("=== gowatch v0.3 ===")
    fmt.Println("")
    for cle, valeur := range metriques {
        fmt.Printf("%-20s : %v\n", cle, valeur)
    }
}

func main() {
    // Déterminer le format souhaité
    format := "texte"
    for _, arg := range os.Args[1:] {
        if arg == "--json" {
            format = "json"
        }
    }

    // Collecter les métriques — on gère l'erreur explicitement
    metriques, err := collecterMetriques()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Erreur : %v\n", err)
        os.Exit(1)
    }

    // Afficher selon le format
    switch format {
    case "json":
        if err := afficherJSON(metriques); err != nil {
            fmt.Fprintf(os.Stderr, "Erreur d'affichage : %v\n", err)
            os.Exit(1)
        }
    default:
        afficherTexte(metriques)
    }
}
```

**Ce que ce refactoring introduit :**
- Des fonctions dédiées avec signature claire (résultat + erreur)
- `fmt.Fprintf(os.Stderr, ...)` pour écrire les erreurs sur le bon flux
- `os.Exit(1)` pour signaler un échec au système appelant
- Le `switch` sur le format — plus propre qu'une chaîne de `if/else`

```bash
go run main.go
go run main.go --json
```

---

## Ce qu'il faut retenir

1. **Go n'a qu'un seul mot-clé de boucle** — `for`. Il couvre tous les cas : compteur, condition, infini, range. Moins à mémoriser, plus à maîtriser.

2. **Les erreurs sont des valeurs, pas des exceptions** — elles se retournent, se propagent manuellement, et se traitent explicitement. C'est verbeux. C'est délibéré. C'est ce qui rend Go si fiable en production.

3. **`defer` est votre gardien des ressources** — ouvrir un fichier ? Connecter une base de données ? Déclarez le `defer` de fermeture immédiatement après. Vous ne l'oublierez plus jamais.

4. **`panic` n'est pas Try/Catch** — c'est la bombe nucléaire de Go. On ne la sort que pour les vrais bugs, jamais pour les erreurs métier normales.

---

## Pour aller plus loin

- 📄 [Error handling and Go](https://go.dev/blog/error-handling-and-go) — Le blog officiel Go sur la philosophie des erreurs
- 📄 [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) — Guide officiel sur ces trois mécanismes
- 📄 [Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors) — Le système moderne d'encapsulation d'erreurs avec `%w`

---

<div align="center">

[⬅️ Chapitre 2.1 — Les briques fondamentales](./01-briques-fondamentales.md) · [👉 Chapitre 2.3 — Composition vs Héritage](./03-composition-vs-heritage.md)

</div>
