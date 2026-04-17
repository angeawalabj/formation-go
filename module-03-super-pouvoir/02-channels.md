# Chapitre 3.2 — Channels : l'art de communiquer

> *"Ne communiquez pas en partageant de la mémoire.
> Partagez de la mémoire en communiquant."*
> — Rob Pike

---

## Le problème

Au chapitre précédent, vous avez lancé trois goroutines qui collectent des métriques en parallèle.

Mais vous avez remarqué le problème : comment récupérer leurs résultats ? Chaque goroutine calcule une valeur — et cette valeur disparaît dans le vide. On a utilisé `fmt.Println` directement dans la goroutine, ce qui est pratique pour un exemple, mais inutilisable dans un vrai programme.

La réponse naïve serait de partager une variable commune :

```go
// ❌ L'approche naïve — et dangereuse
resultats := map[string]float64{}

go func() {
    resultats["CPU"] = 23.4  // Race condition !
}()
go func() {
    resultats["RAM"] = 67.1  // Race condition !
}()
```

Le race detector vous crierait dessus immédiatement. Deux goroutines qui écrivent dans la même map en même temps — c'est une race condition garantie.

La vraie réponse de Go est différente. Radicalement différente.

Plutôt que de partager une variable et de la protéger avec des verrous, **faites passer les données d'une goroutine à l'autre via un canal dédié**.

C'est exactement ce que sont les **channels**.

---

## L'intuition

### La métaphore du tuyau

Un channel en Go, c'est un **tuyau typé** entre deux goroutines.

D'un côté, une goroutine envoie des données dans le tuyau. De l'autre côté, une autre goroutine reçoit ces données. Le tuyau garantit que les données transitent de façon ordonnée et sécurisée — pas de race condition possible, parce que les données n'existent jamais simultanément des deux côtés.

```
Goroutine A ──── envoie ────▶ [channel] ────▶ reçoit ──── Goroutine B
```

### Synchronisation naturelle

Ce qui rend les channels vraiment puissants, c'est qu'ils **synchronisent naturellement** les goroutines.

Par défaut, une goroutine qui envoie dans un channel attend que quelqu'un soit prêt à recevoir. Et une goroutine qui reçoit attend qu'il y ait quelque chose à recevoir.

Ce mécanisme d'attente mutuelle est appelé le **rendez-vous**. Il est gratuit, automatique, et évite une classe entière de bugs de synchronisation.

---

## La solution Go

### Créer et utiliser un channel

```go
package main

import "fmt"

func main() {
    // Créer un channel qui transporte des entiers
    // make(chan TypeDeDonnée)
    ch := make(chan int)

    // Lancer une goroutine qui envoie une valeur
    go func() {
        ch <- 42  // Envoyer 42 dans le channel
                  // Cette goroutine attend que quelqu'un reçoive
    }()

    // Recevoir la valeur depuis le channel
    // main() attend ici jusqu'à ce qu'une valeur soit disponible
    valeur := <-ch

    fmt.Println("Reçu :", valeur)  // Reçu : 42
}
```

> 💡 **La syntaxe `<-`** — La flèche indique la direction du flux de données.
> - `ch <- valeur` : envoyer `valeur` **dans** `ch`
> - `valeur := <-ch` : recevoir depuis `ch` et stocker dans `valeur`
> Lisez-la comme une flèche physique : les données suivent la direction de la flèche.

---

### Channel unbuffered vs buffered

Il existe deux types de channels en Go. La différence est fondamentale.

---

**Channel unbuffered — Le rendez-vous strict**

```go
ch := make(chan int)  // Unbuffered : capacité 0
```

Un channel unbuffered impose un **rendez-vous synchrone** : l'envoi et la réception doivent se produire en même temps. L'émetteur bloque jusqu'à ce qu'un récepteur soit prêt, et vice versa.

```go
ch := make(chan int)

go func() {
    fmt.Println("Goroutine : je vais envoyer...")
    ch <- 100  // Bloque jusqu'à ce que main() reçoive
    fmt.Println("Goroutine : envoi effectué !")
}()

time.Sleep(50 * time.Millisecond)  // Simuler un délai
fmt.Println("Main : je suis prêt à recevoir")
valeur := <-ch  // Débloque la goroutine
fmt.Println("Main : reçu", valeur)
```

**Résultat :**
```
Goroutine : je vais envoyer...
Main : je suis prêt à recevoir
Goroutine : envoi effectué !
Main : reçu 100
```

La goroutine attend bien que `main` soit prêt avant de continuer.

---

**Channel buffered — La file d'attente**

```go
ch := make(chan int, 3)  // Buffered : capacité 3
```

Un channel buffered a une **file d'attente interne**. L'émetteur peut envoyer jusqu'à `capacité` valeurs sans qu'un récepteur soit prêt. Il ne bloque que si la file est pleine.

```go
ch := make(chan string, 3)

// On peut envoyer 3 valeurs sans goroutine réceptrice
ch <- "premier"
ch <- "deuxième"
ch <- "troisième"
// ch <- "quatrième"  // ← Bloquerait : file pleine

fmt.Println(<-ch)  // premier
fmt.Println(<-ch)  // deuxième
fmt.Println(<-ch)  // troisième
```

> 💡 **Quand utiliser lequel ?**
>
> **Unbuffered** — quand vous voulez une synchronisation stricte. Garantit que l'émetteur et le récepteur sont "en phase". Idéal pour signaler qu'une tâche est terminée.
>
> **Buffered** — quand vous voulez découpler l'émetteur du récepteur. Utile quand le producteur est plus rapide que le consommateur par intermittence, ou pour agréger des résultats depuis plusieurs goroutines.

---

### Récupérer les résultats de plusieurs goroutines

Voici le pattern fondamental pour collecter les résultats de plusieurs goroutines en parallèle :

```go
package main

import (
    "fmt"
    "time"
)

type Resultat struct {
    Source string
    Valeur float64
}

func collecter(source string, valeur float64, delai time.Duration, ch chan<- Resultat) {
    // chan<- signifie : ce channel est en ÉCRITURE SEULE dans cette fonction
    time.Sleep(delai)
    ch <- Resultat{Source: source, Valeur: valeur}
}

func main() {
    // Channel buffered de capacité 3 — un slot par goroutine
    resultats := make(chan Resultat, 3)

    // Lancer 3 collectes en parallèle
    go collecter("CPU",       23.4, 50*time.Millisecond, resultats)
    go collecter("RAM",       67.1, 30*time.Millisecond, resultats)
    go collecter("Goroutines", 4.0, 10*time.Millisecond, resultats)

    // Récupérer les 3 résultats (dans l'ordre d'arrivée, pas d'envoi)
    for i := 0; i < 3; i++ {
        r := <-resultats
        fmt.Printf("%-15s : %.2f\n", r.Source, r.Valeur)
    }
}
```

**Résultat :**
```
Goroutines      : 4.00   ← arrive en premier (10ms)
RAM             : 67.10  ← arrive en deuxième (30ms)
CPU             : 23.40  ← arrive en dernier (50ms)
```

> 🔍 **`chan<-` et `<-chan` — Channels directionnels**
>
> Go permet de spécifier la direction d'un channel dans une signature de fonction :
> - `ch chan<- int` : channel en **écriture seule** (la fonction peut seulement envoyer)
> - `ch <-chan int` : channel en **lecture seule** (la fonction peut seulement recevoir)
> - `ch chan int`   : channel **bidirectionnel**
>
> C'est une bonne pratique de toujours utiliser des channels directionnels dans les signatures de fonctions. Le compilateur vérifiera que vous ne faites pas d'erreur de direction.

---

### Fermer un channel — et itérer proprement

Quand une goroutine a fini d'envoyer des données, elle peut **fermer** le channel pour signaler aux récepteurs qu'il n'y aura plus rien à recevoir.

```go
func produire(ch chan<- int) {
    for i := 1; i <= 5; i++ {
        ch <- i
    }
    close(ch)  // Signale : plus rien ne viendra
}

func main() {
    ch := make(chan int, 5)
    go produire(ch)

    // range sur un channel reçoit jusqu'à ce que le channel soit fermé
    for valeur := range ch {
        fmt.Println("Reçu :", valeur)
    }
    fmt.Println("Channel fermé, fin de la boucle")
}
```

**Résultat :**
```
Reçu : 1
Reçu : 2
Reçu : 3
Reçu : 4
Reçu : 5
Channel fermé, fin de la boucle
```

> ⚠️ **Règles sur la fermeture d'un channel**
>
> 1. **Seul l'émetteur ferme** — ne jamais fermer un channel du côté du récepteur
> 2. **Fermer deux fois = panic** — un channel ne se ferme qu'une seule fois
> 3. **Envoyer dans un channel fermé = panic** — vérifiez toujours que c'est bien l'émetteur qui ferme, après son dernier envoi
> 4. **Recevoir depuis un channel fermé est sûr** — on récupère les valeurs restantes, puis la valeur zéro du type indéfiniment

**Détecter si un channel est fermé :**
```go
valeur, ok := <-ch
if !ok {
    fmt.Println("Channel fermé")
}
```

---

### `select` — L'aiguilleur de trafic

`select` permet d'attendre sur **plusieurs channels simultanément** et d'agir sur le premier qui est prêt.

```go
func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)

    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "résultat depuis ch1"
    }()

    go func() {
        time.Sleep(50 * time.Millisecond)
        ch2 <- "résultat depuis ch2"
    }()

    // select attend le premier channel prêt
    select {
    case msg := <-ch1:
        fmt.Println("ch1 gagné :", msg)
    case msg := <-ch2:
        fmt.Println("ch2 gagné :", msg)
    }
}
```

**Résultat :**
```
ch2 gagné : résultat depuis ch2
```

`ch2` arrive en premier (50ms vs 100ms), donc `select` exécute son case.

> 💡 **`select` avec `default`** — Si aucun channel n'est prêt et qu'un case `default` est présent, il s'exécute immédiatement. C'est utile pour faire du **non-blocking receive** :
>
> ```go
> select {
> case msg := <-ch:
>     fmt.Println("Message reçu :", msg)
> default:
>     fmt.Println("Pas de message disponible, je continue")
> }
> ```

---

**`select` avec timeout — Pattern essentiel**

```go
func collecterAvecTimeout(source string, duree time.Duration) (float64, error) {
    resultCh := make(chan float64, 1)

    go func() {
        time.Sleep(duree)  // Simule une collecte lente
        resultCh <- 42.0
    }()

    select {
    case valeur := <-resultCh:
        return valeur, nil

    case <-time.After(100 * time.Millisecond):
        // time.After retourne un channel qui reçoit après le délai indiqué
        return 0, fmt.Errorf("timeout : %s n'a pas répondu dans les délais", source)
    }
}

func main() {
    // Collecte rapide — succès
    val, err := collecterAvecTimeout("CPU", 50*time.Millisecond)
    if err != nil {
        fmt.Println("Erreur :", err)
    } else {
        fmt.Printf("CPU : %.2f\n", val)
    }

    // Collecte lente — timeout
    val, err = collecterAvecTimeout("DisqueLent", 200*time.Millisecond)
    if err != nil {
        fmt.Println("Erreur :", err)
    } else {
        fmt.Printf("DisqueLent : %.2f\n", val)
    }
}
```

**Résultat :**
```
CPU : 42.00
Erreur : timeout : DisqueLent n'a pas répondu dans les délais
```

> 💡 **`time.After`** retourne un `<-chan time.Time` qui envoie une valeur après le délai spécifié. Combiné avec `select`, c'est le moyen le plus idiomatique d'implémenter un timeout en Go.

---

### Le pattern Fan-out / Fan-in

C'est un des patterns de concurrence les plus utilisés en production.

**Fan-out** : distribuer le travail sur plusieurs goroutines.
**Fan-in** : agréger les résultats de plusieurs goroutines dans un seul channel.

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type Metrique struct {
    Source string
    Valeur float64
    Erreur error
}

// fanOut distribue les sources sur des goroutines individuelles
// et retourne un channel unique qui agrège tous les résultats
func fanOut(sources []string) <-chan Metrique {
    // Channel de sortie — toutes les goroutines y écriront
    out := make(chan Metrique, len(sources))

    var wg sync.WaitGroup  // On en parle au chapitre suivant — retenez juste que
                           // wg.Wait() attend que toutes les goroutines finissent

    for _, source := range sources {
        wg.Add(1)
        s := source
        go func() {
            defer wg.Done()
            // Simuler une collecte avec durée variable
            time.Sleep(time.Duration(len(s)) * 10 * time.Millisecond)
            out <- Metrique{Source: s, Valeur: float64(len(s)) * 10.0}
        }()
    }

    // Fermer le channel quand toutes les goroutines ont terminé
    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    sources := []string{"CPU", "RAM", "Disque", "Réseau", "Goroutines"}

    debut := time.Now()

    // Fan-out : chaque source dans sa goroutine
    // Fan-in : tous les résultats dans un seul channel
    resultats := fanOut(sources)

    // Consommer le channel jusqu'à fermeture
    for m := range resultats {
        fmt.Printf("%-12s : %.0f\n", m.Source, m.Valeur)
    }

    fmt.Printf("\nTemps total : %v\n", time.Since(debut))
}
```

**Résultat :**
```
CPU          : 30
RAM          : 30
Disque       : 60
Réseau       : 60
Goroutines   : 100

Temps total : 101ms
(Sans concurrence : 30+30+60+60+100 = 280ms)
```

---

## 🛠️ Projet fil rouge — `gowatch` avec channels

On remplace le `time.Sleep` temporaire par une vraie architecture avec channels.

```go
package main

import (
    "fmt"
    "os"
    "os/signal"
    "runtime"
    "sync"
    "syscall"
    "time"
)

// Metrique représente un résultat de collecte
type Metrique struct {
    Source string
    Valeur float64
    Unite  string
    Erreur error
}

// Source définit une source de métrique
type Source struct {
    Nom      string
    Unite    string
    Collecte func() (float64, error)
}

// collecterSources lance toutes les sources en parallèle
// et retourne un channel fermé automatiquement quand tout est collecté
func collecterSources(sources []Source) <-chan Metrique {
    out := make(chan Metrique, len(sources))
    var wg sync.WaitGroup

    for _, src := range sources {
        wg.Add(1)
        s := src
        go func() {
            defer wg.Done()
            valeur, err := s.Collecte()
            out <- Metrique{
                Source: s.Nom,
                Valeur: valeur,
                Unite:  s.Unite,
                Erreur: err,
            }
        }()
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}

func main() {
    // Définir les sources de métriques
    sources := []Source{
        {
            Nom:   "CPU Logiques",
            Unite: "cœurs",
            Collecte: func() (float64, error) {
                time.Sleep(10 * time.Millisecond)
                return float64(runtime.NumCPU()), nil
            },
        },
        {
            Nom:   "Goroutines",
            Unite: "actives",
            Collecte: func() (float64, error) {
                time.Sleep(5 * time.Millisecond)
                return float64(runtime.NumGoroutine()), nil
            },
        },
        {
            Nom:   "Latence Locale",
            Unite: "ms",
            Collecte: func() (float64, error) {
                debut := time.Now()
                time.Sleep(20 * time.Millisecond) // Simule un ping local
                return float64(time.Since(debut).Milliseconds()), nil
            },
        },
    }

    // Gérer l'arrêt propre avec Ctrl+C
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    fmt.Println("=== gowatch v0.6 — Ctrl+C pour arrêter ===\n")

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    snapshots := 0

    for {
        select {
        case <-ticker.C:
            // Collecter en parallèle à chaque tick
            snapshots++
            debut := time.Now()
            fmt.Printf("[%s]\n", time.Now().Format("15:04:05"))

            for m := range collecterSources(sources) {
                if m.Erreur != nil {
                    fmt.Printf("  %-15s : ERREUR (%v)\n", m.Source, m.Erreur)
                } else {
                    fmt.Printf("  %-15s : %.0f %s\n", m.Source, m.Valeur, m.Unite)
                }
            }
            fmt.Printf("  Collecte en    : %v\n\n", time.Since(debut))

        case <-sigCh:
            fmt.Printf("\nArrêt propre. %d snapshot(s) collecté(s).\n", snapshots)
            os.Exit(0)
        }
    }
}
```

**Lancez et observez :**
```bash
go run main.go
# Laissez tourner quelques secondes, puis Ctrl+C
```

**Résultat :**
```
=== gowatch v0.6 — Ctrl+C pour arrêter ===

[15:42:01]
  CPU Logiques    : 8 cœurs
  Goroutines      : 5 actives
  Latence Locale  : 20 ms
  Collecte en     : 21ms

[15:42:03]
  CPU Logiques    : 8 cœurs
  Goroutines      : 5 actives
  Latence Locale  : 20 ms
  Collecte en     : 21ms

^C
Arrêt propre. 2 snapshot(s) collecté(s).
```

`gowatch` est maintenant un vrai daemon de monitoring : collecte concurrente, rafraîchissement automatique, et arrêt propre sur signal système.

---

## Ce qu'il faut retenir

1. **Un channel est un tuyau typé entre goroutines** — les données transitent de façon ordonnée et sécurisée. Pas de race condition possible sur les données qui passent par un channel.

2. **Unbuffered = rendez-vous synchrone, Buffered = file d'attente** — choisir le bon type dépend du couplage souhaité entre émetteur et récepteur.

3. **`select` attend le premier channel prêt** — combiné avec `time.After`, c'est le mécanisme de timeout idiomatique de Go.

4. **Seul l'émetteur ferme un channel** — et `range` sur un channel itère jusqu'à sa fermeture. Ce pattern est universel en Go concurrent.

5. **Fan-out / Fan-in** — distribuer le travail sur N goroutines et agréger les résultats dans un channel unique. C'est le pattern de concurrence le plus utile du quotidien.

---

## Pour aller plus loin

- 📄 [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines) — Le blog officiel Go sur les pipelines avec channels
- 📄 [Share Memory By Communicating](https://go.dev/blog/codelab-share) — L'article de référence sur la philosophie des channels
- 🎥 [Advanced Go Concurrency Patterns — Sameer Ajmani](https://go.dev/talks/2013/advconc.slide) — Patterns avancés présentés par l'équipe Go

---

<div align="center">

[⬅️ Chapitre 3.1 — Goroutines](./01-goroutines.md) · [👉 Chapitre 3.3 — Patterns de concurrence avancés](./03-patterns-concurrence.md)

</div>
