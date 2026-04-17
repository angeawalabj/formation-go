# Chapitre 3.1 — Goroutines : le chaos ordonné

> *"Un thread OS, c'est un camion 38 tonnes. Une goroutine, c'est un coursier à vélo.
> La ville préfère les vélos."*

---

## Le problème

Votre serveur web reçoit 10 000 requêtes simultanées.

Avec un serveur classique en Python ou Java, chaque requête est traitée par un **thread**. Un thread OS consomme environ **1 à 2 Mo de mémoire** juste pour sa pile d'exécution. 10 000 threads, c'est donc entre 10 et 20 Go de RAM — rien que pour les threads, avant même de traiter quoi que ce soit.

Et créer un thread OS n'est pas gratuit non plus. Le système d'exploitation doit allouer des ressources, initialiser une pile, enregistrer le thread dans son planificateur. C'est une opération coûteuse qui prend des microsecondes — ce qui semble rapide jusqu'à ce que vous en créiez des milliers par seconde.

C'est pourquoi les serveurs haute performance sous Python ou Java utilisent des architectures complexes : pools de threads, event loops, callbacks, promesses, async/await. Des couches d'abstraction empilées pour contourner la limitation fondamentale des threads OS.

Go résout ce problème différemment. Radicalement différemment.

---

## L'intuition

### Thread OS vs Goroutine

Imaginez une ville avec un réseau de livraison.

Un **thread OS**, c'est un camion de livraison 38 tonnes. Il est puissant, peut transporter beaucoup de marchandises, mais il est lourd, coûteux à faire démarrer, et la ville ne peut en faire circuler qu'un nombre limité simultanément — les rues sont saturées, le parking est rare.

Une **goroutine**, c'est un coursier à vélo. Il démarre instantanément, consomme très peu d'espace, peut se faufiler partout. Vous pouvez en avoir des dizaines de milliers dans la même ville sans saturer l'infrastructure.

La ville (votre CPU), c'est le **scheduler de Go** qui décide comment faire circuler efficacement tous ces coursiers sur les routes disponibles (les threads OS, qui eux existent toujours en dessous).

### Le modèle M:N

Go utilise ce qu'on appelle le **multiplexage M:N** : M goroutines s'exécutent sur N threads OS.

```
Goroutines (légères, des milliers)
    ↓  multiplexées par le scheduler Go
Threads OS (lourds, quelques dizaines)
    ↓  gérés par
CPU (quelques cœurs)
```

Le scheduler Go est intégré au runtime — il s'exécute dans votre programme lui-même. Il décide quelle goroutine s'exécute sur quel thread à quel moment, sans que le système d'exploitation ait besoin d'intervenir. C'est ce qui rend le changement de contexte entre goroutines si rapide : il se passe en espace utilisateur, pas en espace noyau.

**Résultat concret :**
- Une goroutine consomme **environ 2 Ko** de mémoire au démarrage (vs 1-2 Mo pour un thread)
- Sa pile **grandit et rétrécit dynamiquement** selon les besoins
- Créer une goroutine prend **quelques microsecondes** (vs des dizaines pour un thread)
- Vous pouvez lancer **des centaines de milliers** de goroutines sur une machine ordinaire

---

## La solution Go

### Le mot-clé `go` — Un mot, une révolution

Lancer une goroutine en Go se fait avec un seul mot-clé : `go`.

```go
package main

import (
    "fmt"
    "time"
)

func direBonjour(nom string) {
    fmt.Printf("Bonjour, %s !\n", nom)
}

func main() {
    // Appel normal — bloquant, séquentiel
    direBonjour("Alice")

    // Appel concurrent — non bloquant, en arrière-plan
    go direBonjour("Bob")
    go direBonjour("Carol")
    go direBonjour("David")

    // ATTENTION : sans cette pause, main() se termine
    // avant que les goroutines aient le temps de s'exécuter
    time.Sleep(100 * time.Millisecond)

    fmt.Println("Terminé.")
}
```

**Résultat possible :**
```
Bonjour, Alice !
Bonjour, Carol !
Bonjour, Bob !
Bonjour, David !
Terminé.
```

> ⚠️ **Attention** — L'ordre d'exécution des goroutines n'est **pas garanti**. Bob, Carol et David peuvent s'afficher dans n'importe quel ordre à chaque exécution. C'est la nature de la concurrence — et c'est normal. Votre code ne doit jamais dépendre d'un ordre précis d'exécution entre goroutines.

> ⚠️ **Le piège du `time.Sleep`** — Utiliser `time.Sleep` pour attendre des goroutines est une mauvaise pratique. C'est fragile et imprévisible. On le fait ici pour simplifier l'exemple. Le Module 03 vous donnera les vrais outils : `WaitGroup` et `channels`.

---

### Goroutines avec des fonctions anonymes

On peut aussi lancer une goroutine avec une fonction définie à la volée :

```go
func main() {
    message := "Bonjour depuis une goroutine anonyme"

    go func() {
        fmt.Println(message)
    }()  // Les () à la fin appellent immédiatement la fonction

    time.Sleep(100 * time.Millisecond)
}
```

**Variante avec paramètre — la bonne pratique**

```go
func main() {
    for i := 0; i < 5; i++ {
        // ✅ Passer i en paramètre — chaque goroutine a sa propre copie
        go func(numero int) {
            fmt.Printf("Goroutine numéro %d\n", numero)
        }(i)
    }

    time.Sleep(100 * time.Millisecond)
}
```

> ⚠️ **Le piège classique des closures en boucle**
>
> ```go
> // ❌ MAUVAIS — toutes les goroutines partagent la même variable i
> for i := 0; i < 5; i++ {
>     go func() {
>         fmt.Println(i)  // i peut valoir 5 pour toutes les goroutines !
>     }()
> }
>
> // ✅ BON — chaque goroutine reçoit sa propre copie de i
> for i := 0; i < 5; i++ {
>     go func(n int) {
>         fmt.Println(n)
>     }(i)
> }
> ```
>
> C'est un des bugs les plus fréquents chez les débutants en concurrence Go. La variable `i` est partagée entre toutes les goroutines dans la version incorrecte — au moment où une goroutine s'exécute, `i` a peut-être déjà changé de valeur.

---

### Le scheduler Go — Comment ça marche vraiment

Le scheduler Go utilise un algorithme appelé **work-stealing**. Voici l'intuition :

Imaginez plusieurs caisses dans un supermarché (les threads OS). Chaque caisse a sa propre file d'attente de clients (les goroutines). Quand une caisse finit de servir ses clients avant les autres, au lieu de rester inactive, elle "vole" des clients dans la file d'une caisse plus chargée.

C'est ce que fait le scheduler Go : les threads sous-utilisés volent des goroutines aux threads surchargés. Résultat — les CPU restent occupés tant qu'il y a du travail à faire.

```go
import "runtime"

func main() {
    // Voir combien de threads OS Go peut utiliser
    // Par défaut : égal au nombre de CPU logiques de la machine
    fmt.Println("GOMAXPROCS :", runtime.GOMAXPROCS(0))

    // Forcer l'utilisation d'un seul thread (mode coopératif)
    // Rarement utile en pratique, mais instructif pour comprendre
    runtime.GOMAXPROCS(1)
}
```

> 💡 **`GOMAXPROCS`** — Cette variable contrôle combien de threads OS le scheduler Go peut utiliser simultanément. Sa valeur par défaut est le nombre de CPU logiques de votre machine. En production, vous ne la changez presque jamais — la valeur par défaut est optimale dans 99% des cas.

---

### Le Race Detector — Votre meilleur ami

Une **race condition** se produit quand deux goroutines accèdent à la même variable en même temps, et qu'au moins l'une d'elles la modifie. Le résultat est imprévisible et le bug est extrêmement difficile à reproduire.

```go
// ⚠️ Race condition — NE PAS FAIRE
package main

import (
    "fmt"
    "sync"
)

func main() {
    compteur := 0
    var wg sync.WaitGroup

    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            compteur++  // ← RACE CONDITION : plusieurs goroutines écrivent en même temps
        }()
    }

    wg.Wait()
    fmt.Println("Compteur :", compteur)
    // Résultat imprévisible : peut être 998, 1000, ou n'importe quoi entre les deux
}
```

Go intègre un **détecteur de race conditions** que vous activez avec le flag `-race` :

```bash
go run -race main.go
# ou
go build -race -o monprogramme main.go
```

Quand une race condition est détectée, Go l'affiche clairement :

```
==================
WARNING: DATA RACE
Write at 0x00c0000b4010 by goroutine 7:
  main.main.func1()
      /home/user/main.go:16 +0x30

Previous write at 0x00c0000b4010 by goroutine 6:
  main.main.func1()
      /home/user/main.go:16 +0x30
==================
```

> 💡 **Bonne pratique** — Activez `-race` systématiquement pendant le développement et dans vos tests automatisés. Ne le désactivez en production que si les performances sont critiques — le mode race ajoute environ 5 à 10x de surcharge. La solution à une race condition, vous la verrez au Chapitre 3.3 avec les Mutex et les channels.

---

### Le Goroutine Leak — L'autre piège majeur

Une **goroutine leak** se produit quand une goroutine est lancée mais ne se termine jamais — elle reste en mémoire indéfiniment, consommant des ressources.

```go
// ⚠️ Goroutine leak — NE PAS FAIRE
func lancerTache() {
    go func() {
        for {
            // Cette goroutine tourne indéfiniment
            // Si personne ne l'arrête, elle ne s'arrêtera jamais
            traiterDonnees()
            time.Sleep(time.Second)
        }
    }()
    // La fonction retourne, mais la goroutine continue de tourner
}
```

Les goroutine leaks sont insidieuses : elles n'apparaissent pas immédiatement. Elles s'accumulent silencieusement jusqu'à ce que votre programme consomme trop de mémoire et soit tué par le système.

**Comment les détecter :**

```go
import "runtime"

// Afficher le nombre de goroutines actives
fmt.Println("Goroutines actives :", runtime.NumGoroutine())
```

Si ce nombre ne cesse de croître pendant l'exécution de votre programme, vous avez probablement une goroutine leak. La solution — le `Context` — sera vue au Chapitre 3.3.

---

## 🛠️ Projet fil rouge — `gowatch` lance ses premières goroutines

On refactorise la collecte de métriques pour qu'elle soit **concurrente** : chaque source de données tournera dans sa propre goroutine.

Pour l'instant, on utilise `time.Sleep` pour attendre — on le remplacera par des channels au prochain chapitre.

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

// Source représente une source de métriques
type Source struct {
    Nom      string
    Collecte func() float64
}

// collecterDepuisSource simule la collecte d'une métrique
// Dans un vrai système, elle lirait /proc/stat, /proc/meminfo, etc.
func collecterDepuisSource(source Source) {
    debut := time.Now()
    valeur := source.Collecte()
    duree := time.Since(debut)

    fmt.Printf("[%s] Valeur=%.2f  Durée=%v  Goroutines=%d\n",
        source.Nom,
        valeur,
        duree,
        runtime.NumGoroutine(),
    )
}

func main() {
    // Définir nos sources de métriques
    sources := []Source{
        {
            Nom: "CPU",
            Collecte: func() float64 {
                // Simulation d'une collecte qui prend du temps
                time.Sleep(50 * time.Millisecond)
                return 23.4
            },
        },
        {
            Nom: "RAM",
            Collecte: func() float64 {
                time.Sleep(30 * time.Millisecond)
                return 67.1
            },
        },
        {
            Nom: "Goroutines",
            Collecte: func() float64 {
                time.Sleep(10 * time.Millisecond)
                return float64(runtime.NumGoroutine())
            },
        },
    }

    fmt.Println("=== gowatch v0.5 — Collecte concurrente ===")
    fmt.Printf("Goroutines au démarrage : %d\n\n", runtime.NumGoroutine())

    debut := time.Now()

    // Lancer chaque collecte dans sa propre goroutine
    for _, source := range sources {
        s := source // ✅ Copie locale pour éviter le piège des closures
        go collecterDepuisSource(s)
    }

    // ⚠️ Temporaire — on remplacera par WaitGroup + channels au prochain chapitre
    time.Sleep(200 * time.Millisecond)

    fmt.Printf("\nTemps total (concurrent) : %v\n", time.Since(debut))
    fmt.Println("(Sans concurrence, ce serait ~90ms = 50+30+10)")
}
```

**Lancez et observez :**
```bash
go run main.go
```

**Résultat attendu :**
```
=== gowatch v0.5 — Collecte concurrente ===
Goroutines au démarrage : 1

[Goroutines] Valeur=4.00  Durée=10ms  Goroutines=4
[RAM]        Valeur=67.10 Durée=30ms  Goroutines=4
[CPU]        Valeur=23.40 Durée=50ms  Goroutines=4

Temps total (concurrent) : 201ms
(Sans concurrence, ce serait ~90ms = 50+30+10)
```

> 💡 **Observez le gain** — Les trois sources tournent en parallèle. La durée totale est celle de la source la plus lente (CPU à 50ms), pas la somme des trois (90ms). C'est l'essence de la concurrence.

**Lancez avec le race detector :**
```bash
go run -race main.go
```
Pas de race condition ici — chaque goroutine travaille sur ses propres données locales. Au chapitre suivant, quand on partagera des données entre goroutines, on verra comment les channels protègent naturellement contre ces problèmes.

---

## Ce qu'il faut retenir

1. **`go maFonction()` lance une goroutine** — une tâche légère, concurrente, qui s'exécute en arrière-plan. Un seul mot-clé. Pas de configuration, pas de pool à gérer.

2. **Une goroutine coûte ~2 Ko de mémoire** — contre 1-2 Mo pour un thread OS. Vous pouvez en avoir des centaines de milliers sans saturer votre machine.

3. **L'ordre d'exécution n'est pas garanti** — votre code concurrent ne doit jamais supposer qu'une goroutine s'exécute avant une autre.

4. **Passez les variables en paramètre dans les boucles** — le piège de la closure qui capture `i` mutable est le bug numéro un des débutants en concurrence Go.

5. **Utilisez `-race` pendant le développement** — le race detector est gratuit à activer et peut vous éviter des heures de débogage en production.

6. **`time.Sleep` n'est pas une synchronisation** — c'est un pansement temporaire. Les vrais outils arrivent aux chapitres suivants.

---

## Pour aller plus loin

- 📄 [Goroutines — A Tour of Go](https://go.dev/tour/concurrency/1) — L'introduction interactive officielle
- 📄 [Go Concurrency Patterns — Rob Pike](https://go.dev/talks/2012/concurrency.slide) — La présentation de référence
- 🎥 [Concurrency is not Parallelism — Rob Pike](https://go.dev/blog/waza-talk) — La distinction fondamentale entre concurrence et parallélisme

---

<div align="center">

[⬅️ Retour au Module 03](./README.md) · [👉 Chapitre 3.2 — Channels : l'art de communiquer](./02-channels.md)

</div>
