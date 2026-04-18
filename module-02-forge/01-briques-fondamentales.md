# Chapitre 2.1 — Les briques fondamentales

> *"Un langage qui ne change pas votre façon de penser ne vaut pas la peine d'être appris."*
> — Alan Perlis

---

## Le problème

Vous avez déjà déclaré des variables dans un autre langage. Vous savez ce qu'est un tableau. Vous avez probablement utilisé un dictionnaire ou un objet pour stocker des données clé-valeur.

Alors pourquoi consacrer un chapitre entier aux "briques fondamentales" de Go ?

Parce que Go fait certaines choses différemment. Pas de façon arbitraire — mais de façon délibérée, avec des raisons précises. Et si vous ne comprenez pas ces raisons, vous allez passer les prochaines semaines à écrire du "Python avec une syntaxe Go" — du code qui marche, mais qui ne tire pas parti de ce que le langage offre vraiment.

Ce chapitre est court. Mais lisez-le attentivement. Ce sont les fondations sur lesquelles tout le reste repose.

---

## L'intuition

Imaginez que vous construisez une maison.

En Python, les briques sont en argile souple — elles s'adaptent à presque n'importe quelle forme, elles sont faciles à manipuler, mais elles ne supportent pas les charges lourdes sans renfort.

En C++, les briques sont en acier brut — extrêmement solides, mais vous devez les couper vous-même, les assembler vous-même, et gérer vous-même les déchets de construction.

En Go, les briques sont en béton préfabriqué — solides, de taille fixe, faciles à assembler, et livrées avec un système de nettoyage automatique. Vous ne les façonnez pas à la main, mais vous savez exactement combien elles pèsent et combien de charge elles supportent.

Cette métaphore va guider tout ce chapitre.

---

## La solution Go

### Les variables — Ce que vous savez déjà, mais en mieux

En Go, il y a deux façons de déclarer une variable. Choisir la bonne selon le contexte est votre premier geste idiomatique.

**Forme longue — avec `var`**
```go
var nom string = "Alice"
var age  int    = 30
var actif bool  = true
```

**Forme courte — avec `:=` (la plus utilisée)**
```go
nom   := "Alice"  // Go infère le type string
age   := 30       // Go infère le type int
actif := true     // Go infère le type bool
```

L'opérateur `:=` déclare et initialise la variable en même temps. Go déduit le type automatiquement à partir de la valeur. C'est ce qu'on appelle l'**inférence de type**.

> 💡 **Astuce** — Utilisez `:=` à l'intérieur des fonctions (c'est la norme). Utilisez `var` au niveau du package (hors des fonctions) ou quand vous voulez déclarer une variable sans l'initialiser immédiatement.

```go
// ✅ Déclaration sans initialisation — var obligatoire ici
var compteur int  // valeur par défaut : 0
var message string  // valeur par défaut : ""
var actif bool      // valeur par défaut : false
```

> ⚠️ **Attention — La règle d'or de Go** : toute variable déclarée **doit** être utilisée. Si vous déclarez `age := 30` et ne l'utilisez jamais dans votre code, Go refusera de compiler. C'est intentionnel — Go déteste le code mort.

---

### Les types de base — Le béton préfabriqué

Go est un langage à **typage statique fort**. Chaque variable a un type défini à la compilation, et ce type ne change pas.

Voici les types que vous utiliserez au quotidien :

```go
// Entiers
var a int     = 42      // Taille dépend de l'OS (32 ou 64 bits)
var b int64   = 42      // Toujours 64 bits
var c int32   = 42      // Toujours 32 bits
var d uint    = 42      // Entier non signé (pas de négatif)

// Flottants
var e float64 = 3.14    // Le plus utilisé
var f float32 = 3.14    // Moins précis, rarement nécessaire

// Texte
var g string  = "Bonjour"   // Toujours entre guillemets doubles

// Booléen
var h bool    = true

// Byte — utile pour manipuler des données binaires
var i byte    = 65      // 'A' en ASCII
```

> 🔍 **Zoom profil — Développeurs Python**
> En Python, `42` peut devenir `42.0` ou `"42"` selon le contexte. En Go, `42` est un `int`, et le rester. Vous ne pouvez pas additionner un `int` et un `float64` sans conversion explicite. Au début c'est agaçant. Très vite, c'est rassurant — vous savez exactement ce que manipule votre code.

```go
// ⚠️ Ceci ne compile pas en Go
x := 10
y := 3.14
// z := x + y  // ERREUR : impossible d'additionner int et float64

// ✅ Il faut convertir explicitement
z := float64(x) + y  // z = 13.14
```

---

### Les Slices — Oubliez les tableaux classiques

En Go, il existe les **tableaux** (arrays) et les **slices**. Dans 95% des cas, vous utiliserez des slices. Voici pourquoi.

Un **tableau** en Go a une taille fixe définie à la compilation :
```go
// Tableau de 5 entiers — taille fixe, immuable
var tableau [5]int = [5]int{10, 20, 30, 40, 50}
```

Un **slice** est une vue dynamique sur un tableau sous-jacent. Sa taille peut varier :
```go
// Slice d'entiers — taille dynamique
nombres := []int{10, 20, 30, 40, 50}

// Ajouter un élément
nombres = append(nombres, 60)

// Accéder à un élément
fmt.Println(nombres[0])  // 10

// Longueur actuelle
fmt.Println(len(nombres))  // 6

// Sous-slice (comme Python : nombres[1:3])
sousEnsemble := nombres[1:3]  // [20, 30]
```

**Pourquoi les slices sont supérieurs aux tableaux classiques :**

```go
// Créer un slice vide, prêt à être rempli
villes := []string{}

// Ajouter des éléments dynamiquement
villes = append(villes, "Paris")
villes = append(villes, "Dakar")
villes = append(villes, "Montréal")

fmt.Println(villes)         // [Paris Dakar Montréal]
fmt.Println(len(villes))    // 3
```

> 💡 **Astuce — `make` pour les slices de grande taille**
> Si vous savez à l'avance combien d'éléments vous allez stocker, utilisez `make` pour pré-allouer la mémoire. C'est plus performant qu'une série d'`append` :
>
> ```go
> // Crée un slice de 0 éléments avec une capacité de 1000
> donnees := make([]int, 0, 1000)
> ```

---

### Itérer sur un Slice — Le `range`

Pour parcourir un slice, Go utilise `range` — une construction qui retourne à la fois l'index et la valeur :

```go
fruits := []string{"mangue", "papaye", "ananas"}

for index, fruit := range fruits {
    fmt.Printf("%d : %s\n", index, fruit)
}
// 0 : mangue
// 1 : papaye
// 2 : ananas
```

Si vous n'avez pas besoin de l'index, utilisez `_` pour l'ignorer explicitement :

```go
for _, fruit := range fruits {
    fmt.Println(fruit)
}
```

> ⚠️ **Attention** — En Go, on ne peut pas ignorer silencieusement une valeur de retour. Si `range` retourne deux valeurs et que vous n'en utilisez qu'une, vous **devez** utiliser `_` pour l'autre. C'est encore cette règle : rien n'est caché, tout est explicite.

---

### Les Maps — Le dictionnaire haute performance

Une `map` en Go est l'équivalent d'un dictionnaire Python ou d'un objet JavaScript. C'est une structure clé-valeur :

```go
// Déclarer et initialiser une map
capitales := map[string]string{
    "France":   "Paris",
    "Sénégal":  "Dakar",
    "Japon":    "Tokyo",
}

// Accéder à une valeur
fmt.Println(capitales["France"])  // Paris

// Ajouter ou modifier une entrée
capitales["Canada"] = "Ottawa"

// Supprimer une entrée
delete(capitales, "Japon")

// Vérifier si une clé existe — IMPORTANT
capitale, existe := capitales["Allemagne"]
if existe {
    fmt.Println("Capitale :", capitale)
} else {
    fmt.Println("Pays non trouvé")
}
```

> ⚠️ **Attention — Le piège classique des Maps**
> Si vous accédez à une clé qui n'existe pas, Go ne plante pas. Il retourne la **valeur zéro** du type de valeur. Pour un `string`, c'est `""`. Pour un `int`, c'est `0`. Ce comportement silencieux peut créer des bugs difficiles à détecter.
>
> **Toujours vérifier l'existence d'une clé** avec la forme à deux valeurs de retour :
> ```go
> valeur, ok := maMap["cle"]
> if !ok {
>     // La clé n'existe pas
> }
> ```

---

### Créer une Map avec `make`

```go
// Créer une map vide (obligatoire avant d'écrire dedans)
scores := make(map[string]int)

scores["Alice"] = 95
scores["Bob"]   = 87
scores["Carol"] = 92

// Itérer sur une map
for nom, score := range scores {
    fmt.Printf("%s : %d\n", nom, score)
}
```

> 💡 **Astuce** — L'ordre d'itération sur une map en Go est **intentionnellement aléatoire**. Go randomise l'ordre à chaque exécution pour éviter que les développeurs dépendent d'un ordre qui n'est pas garanti. Si vous avez besoin d'un ordre précis, extrayez les clés dans un slice et triez-le.

---

## 🛠️ Projet fil rouge — `gowatch` collecte ses premières métriques

On applique ce qu'on vient d'apprendre à `gowatch`. On va structurer nos premières données de monitoring.

Mettez à jour votre `main.go` :

```go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "runtime"
)

func main() {
    // On collecte les métriques dans une map
    // map[string]interface{} signifie : clés string, valeurs de n'importe quel type
    metriques := map[string]interface{}{
        "systeme":    runtime.GOOS + "/" + runtime.GOARCH,
        "go_version": runtime.Version(),
        "cpu_count":  runtime.NumCPU(),
        "goroutines": runtime.NumGoroutine(),
    }

    // Vérifier si l'argument --json est passé
    format := "texte"
    for _, arg := range os.Args[1:] {
        if arg == "--json" {
            format = "json"
        }
    }

    // Afficher selon le format demandé
    if format == "json" {
        // json.MarshalIndent retourne deux valeurs : le résultat ET une erreur
        donnees, err := json.MarshalIndent(metriques, "", "  ")
        if err != nil {
            fmt.Println("Erreur de sérialisation JSON :", err)
            return
        }
        fmt.Println(string(donnees))
    } else {
        fmt.Println("=== gowatch v0.2 ===")
        fmt.Println("")
        for cle, valeur := range metriques {
            fmt.Printf("%-20s : %v\n", cle, valeur)
        }
    }
}
```

**Testez les deux formats :**

```bash
# Format texte (par défaut)
go run main.go

# Format JSON
go run main.go --json
```

**Résultat en mode JSON :**
```json
{
  "cpu_count": 8,
  "go_version": "go1.23.0",
  "goroutines": 1,
  "systeme": "linux/amd64"
}
```

Vous venez d'utiliser des Maps, des Slices (via `os.Args`), du `range`, et une gestion d'erreur basique — le tout dans un vrai programme utile. Le Module 02 est bien lancé.

---

## Ce qu'il faut retenir

1. **`:=` est votre opérateur de déclaration principal** à l'intérieur des fonctions. `var` est réservé aux déclarations de package ou aux cas sans initialisation immédiate.

2. **Utilisez des Slices, pas des tableaux**. Les tableaux ont une taille fixe et sont rarement nécessaires directement. Les Slices sont dynamiques, idiomatiques, et universels.

3. **Vérifiez toujours l'existence d'une clé dans une Map** avec la forme à deux retours `valeur, ok := map[cle]`. Ne faites jamais confiance à la valeur zéro silencieuse.

4. **Go ne vous laisse pas ignorer du code**. Variables non utilisées, imports inutiles, valeurs de retour silencieuses — tout est une erreur de compilation. C'est une contrainte qui vous protège.

---

## Pour aller plus loin

- 📄 [Go Slices: usage and internals](https://go.dev/blog/slices-intro) — Le blog officiel Go, explique comment les Slices fonctionnent en mémoire
- 📄 [Go Maps in action](https://go.dev/blog/maps) — Guide officiel sur les Maps
- 🔧 [Go Playground](https://go.dev/play/) — Testez du code Go directement dans le navigateur, sans rien installer

---

<div align="center">

[⬅️ Retour au Module 02](./README.md) · [👉 Chapitre 2.2 — La logique sans fioritures](./02-logique-sans-fioritures.md)

</div>
