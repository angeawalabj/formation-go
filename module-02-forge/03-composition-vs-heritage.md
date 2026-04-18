# Chapitre 2.3 — Composition vs Héritage

> *"Favorisez la composition sur l'héritage."*
> — Gang of Four, Design Patterns (1994)
>
> *"Go a juste pris ça au sérieux."*
> — La communauté Go

---

## Le problème

Si vous venez de la programmation orientée objet — Java, Python, PHP, C++ — vous avez appris à modéliser le monde avec des **classes** et de l**'héritage**.

Un `Chien` hérite d'`Animal`. Un `CompteBancaire` hérite de `CompteFinancier`. Une `VoitureElectrique` hérite de `Voiture` qui hérite de `VehiculeMoteur`.

Ça semble logique. Et pendant longtemps, c'était la façon dominante de penser la structure d'un programme.

Mais l'héritage a un problème bien documenté : il **couple** les classes de façon rigide. Modifier une classe parente peut casser silencieusement toutes ses classes filles. Les hiérarchies profondes deviennent impossibles à naviguer. Et le fameux problème du "diamant" — quand une classe hérite de deux parents qui partagent un ancêtre commun — donne des maux de tête à toute une génération de développeurs.

Go a tranché : **pas de classes. Pas d'héritage.**

À la place : des **Structs**, des **méthodes**, et des **interfaces**. Trois outils simples qui, ensemble, permettent de faire tout ce que la POO promet — sans ses défauts.

---

## L'intuition

### L'héritage, c'est quoi vraiment ?

L'héritage dit : *"Un Chien EST UN Animal."*

La composition dit : *"Un Chien A DES comportements : il peut manger, courir, aboyer."*

La différence est subtile mais fondamentale. L'héritage définit ce qu'une chose **est**. La composition définit ce qu'une chose **peut faire**.

Go choisit la composition. Toujours.

### L'analogie du Lego

Pensez aux Lego.

Chaque brique est indépendante. Elle n'hérite de rien. Mais vous pouvez assembler des briques pour créer n'importe quelle structure, aussi complexe soit-elle. Et si une brique ne vous convient plus, vous la retirez sans démolir le reste.

C'est exactement ce que Go propose avec ses Structs et ses interfaces.

---

## La solution Go

### Les Structs — Modéliser des données sans classes

Une `struct` est un type composite qui regroupe des champs de différents types sous un même nom.

```go
// Définir une struct
type Metrique struct {
    Nom    string
    Valeur float64
    Unite  string
}

// Créer une instance
cpu := Metrique{
    Nom:    "CPU",
    Valeur: 23.5,
    Unite:  "%",
}

// Accéder aux champs
fmt.Println(cpu.Nom)    // CPU
fmt.Println(cpu.Valeur) // 23.5
fmt.Println(cpu.Unite)  // %
```

> 💡 **Convention Go** — Les noms de types commencent par une **majuscule** (`Metrique`, pas `metrique`). En Go, la majuscule signifie que l'élément est **exporté** — visible depuis d'autres packages. La minuscule le rend privé au package courant. C'est le système de visibilité de Go : simple, visible, sans mot-clé `public` ou `private`.

---

**Structs imbriquées**

```go
type SystemeInfo struct {
    OS          string
    Architecture string
    CPUCount    int
}

type Snapshot struct {
    Timestamp  time.Time
    Systeme    SystemeInfo   // Struct imbriquée
    Metriques  []Metrique    // Slice de Structs
}

// Créer un snapshot complet
snap := Snapshot{
    Timestamp: time.Now(),
    Systeme: SystemeInfo{
        OS:           runtime.GOOS,
        Architecture: runtime.GOARCH,
        CPUCount:     runtime.NumCPU(),
    },
    Metriques: []Metrique{
        {Nom: "CPU", Valeur: 23.5, Unite: "%"},
        {Nom: "RAM", Valeur: 67.2, Unite: "%"},
    },
}

fmt.Println(snap.Systeme.OS)              // linux
fmt.Println(snap.Metriques[0].Valeur)    // 23.5
```

---

### Les méthodes — Attacher du comportement à une Struct

En Go, une méthode est une fonction attachée à un type. La syntaxe est légèrement différente d'une fonction normale : on ajoute un **récepteur** entre `func` et le nom de la méthode.

```go
type Metrique struct {
    Nom    string
    Valeur float64
    Unite  string
}

// Méthode sur Metrique — le récepteur est (m Metrique)
func (m Metrique) Afficher() {
    fmt.Printf("%s : %.2f %s\n", m.Nom, m.Valeur, m.Unite)
}

// Utilisation
cpu := Metrique{Nom: "CPU", Valeur: 23.5, Unite: "%"}
cpu.Afficher()  // CPU : 23.50 %
```

---

**Récepteur par valeur vs récepteur par pointeur**

C'est une distinction importante en Go.

```go
// Récepteur par VALEUR — travaille sur une COPIE de la struct
// À utiliser quand la méthode ne modifie pas la struct
func (m Metrique) Afficher() {
    fmt.Printf("%s : %.2f %s\n", m.Nom, m.Valeur, m.Unite)
}

// Récepteur par POINTEUR — travaille sur l'ORIGINAL
// À utiliser quand la méthode doit MODIFIER la struct
func (m *Metrique) Mettre_a_jour(nouvelleValeur float64) {
    m.Valeur = nouvelleValeur  // Modifie la vraie struct, pas une copie
}

cpu := Metrique{Nom: "CPU", Valeur: 23.5, Unite: "%"}
cpu.Afficher()              // CPU : 23.50 %
cpu.Mettre_a_jour(41.8)
cpu.Afficher()              // CPU : 41.80 %
```

> 💡 **Règle simple** — Si votre méthode modifie la struct, utilisez un pointeur (`*Metrique`). Si elle se contente de lire, utilisez une valeur (`Metrique`). En cas de doute sur la performance (struct très grande), préférez le pointeur pour éviter la copie.

> 🔍 **Zoom profil — Développeurs Java/Python**
> Le récepteur `(m Metrique)` est l'équivalent de `this` en Java ou `self` en Python. La différence : en Go, vous le nommez vous-même (par convention, une ou deux lettres — la première lettre du type), et vous choisissez explicitement si c'est une copie ou une référence.

---

### Les interfaces — La flexibilité sans couplage

Une interface en Go définit un **ensemble de méthodes** qu'un type doit implémenter. Jusque-là, rien d'inhabituel par rapport à Java ou C#.

La différence fondamentale : en Go, l'implémentation est **implicite**.

```go
// Définir une interface
type Affichable interface {
    Afficher()
}
```

En Java, pour implémenter une interface, vous écrivez `class MaClasse implements MonInterface`. En Go, il n'y a rien à écrire. Si votre type possède la méthode `Afficher()`, il implémente automatiquement l'interface `Affichable`. Point.

C'est ce qu'on appelle le **Duck Typing statique** : *"Si ça marche comme un canard et que ça cancane comme un canard, c'est un canard."* Mais vérifié à la compilation, pas à l'exécution.

```go
type Metrique struct {
    Nom    string
    Valeur float64
    Unite  string
}

// Metrique implémente Affichable SANS le déclarer explicitement
func (m Metrique) Afficher() {
    fmt.Printf("[Métrique] %s : %.2f %s\n", m.Nom, m.Valeur, m.Unite)
}

type AlerteSysteme struct {
    Message  string
    Severite string
}

// AlerteSysteme implémente aussi Affichable SANS le déclarer
func (a AlerteSysteme) Afficher() {
    fmt.Printf("[Alerte %s] %s\n", a.Severite, a.Message)
}

// Une fonction qui accepte N'IMPORTE QUEL type Affichable
func afficherTout(elements []Affichable) {
    for _, element := range elements {
        element.Afficher()
    }
}

func main() {
    elements := []Affichable{
        Metrique{Nom: "CPU", Valeur: 87.3, Unite: "%"},
        AlerteSysteme{Message: "RAM critique", Severite: "HAUTE"},
        Metrique{Nom: "Disque", Valeur: 94.1, Unite: "%"},
    }

    afficherTout(elements)
}
```

**Résultat :**
```
[Métrique] CPU : 87.30 %
[Alerte HAUTE] RAM critique
[Métrique] Disque : 94.10 %
```

> ⚠️ **Le pouvoir de cette approche** — La fonction `afficherTout` ne connaît ni `Metrique` ni `AlerteSysteme`. Elle ne connaît que l'interface `Affichable`. Demain, si vous créez un nouveau type `RapportQuotidien` avec une méthode `Afficher()`, il fonctionnera automatiquement avec `afficherTout` — sans modifier une seule ligne de code existante.

---

**L'interface vide — `interface{}`**

Go a un type spécial : l'interface vide `interface{}` (ou `any` depuis Go 1.18). Elle est implémentée par **tous** les types, car elle ne requiert aucune méthode.

```go
// Accepte n'importe quel type
func afficherNimporteQuoi(valeur interface{}) {
    fmt.Printf("Type : %T, Valeur : %v\n", valeur, valeur)
}

afficherNimporteQuoi(42)          // Type : int, Valeur : 42
afficherNimporteQuoi("bonjour")   // Type : string, Valeur : bonjour
afficherNimporteQuoi(true)        // Type : bool, Valeur : true
```

> ⚠️ **Attention** — L'interface vide est puissante mais dangereuse si overutilisée. Elle désactive les vérifications de type du compilateur. Utilisez-la uniquement quand vous avez vraiment besoin d'accepter n'importe quel type — comme dans les fonctions de sérialisation JSON. Dans tous les autres cas, préférez une interface avec des méthodes explicites.

---

### L'Embedding — La composition en action

Go n'a pas d'héritage, mais il a l'**embedding** : la possibilité d'intégrer un type dans un autre pour réutiliser ses méthodes directement.

```go
type Base struct {
    ID        int
    CreeLe    time.Time
}

// Base a une méthode
func (b Base) Identifier() string {
    return fmt.Sprintf("ID:%d créé le %s", b.ID, b.CreeLe.Format("2006-01-02"))
}

// Metrique intègre Base — elle "hérite" de ses méthodes
type MetriqueComplete struct {
    Base              // Embedding — pas de nom de champ, juste le type
    Nom    string
    Valeur float64
}

func main() {
    m := MetriqueComplete{
        Base:   Base{ID: 1, CreeLe: time.Now()},
        Nom:    "CPU",
        Valeur: 45.2,
    }

    // On accède directement aux méthodes de Base
    fmt.Println(m.Identifier())  // ID:1 créé le 2024-01-15
    fmt.Println(m.Nom)           // CPU
    fmt.Println(m.Base.ID)       // 1 — accès explicite aussi possible
}
```

> 💡 **Embedding vs Héritage** — L'embedding ressemble à de l'héritage mais ce n'en est pas. `MetriqueComplete` ne **devient pas** un `Base`. Elle **contient** un `Base` et en réutilise les méthodes directement. La nuance est importante : si une fonction attend un `Base`, vous ne pouvez pas lui passer un `MetriqueComplete`. C'est la composition, pas l'héritage.

---

### Les interfaces standards de Go — La bibliothèque qui s'adapte à vous

La puissance des interfaces implicites brille dans la bibliothèque standard de Go. Voici les deux plus importantes :

**`fmt.Stringer` — Contrôler l'affichage d'un type**

```go
// Si votre type implémente String() string,
// fmt.Println l'utilisera automatiquement
type Metrique struct {
    Nom    string
    Valeur float64
    Unite  string
}

func (m Metrique) String() string {
    return fmt.Sprintf("%s=%.2f%s", m.Nom, m.Valeur, m.Unite)
}

cpu := Metrique{Nom: "CPU", Valeur: 23.5, Unite: "%"}
fmt.Println(cpu)  // CPU=23.50%  ← utilise automatiquement String()
```

**`io.Reader` et `io.Writer` — L'universalité des flux**

```go
// io.Writer est défini ainsi dans la bibliothèque standard :
type Writer interface {
    Write(p []byte) (n int, err error)
}

// Un fichier l'implémente. Un buffer mémoire l'implémente.
// Une connexion réseau l'implémente. Votre propre type peut l'implémenter.
// Et toutes les fonctions qui acceptent un io.Writer fonctionnent avec tous.

func ecrireDonnees(w io.Writer, donnees string) error {
    _, err := fmt.Fprintln(w, donnees)
    return err
}

// Écrire dans un fichier
fichier, _ := os.Create("sortie.txt")
ecrireDonnees(fichier, "Bonjour depuis un fichier")

// Écrire dans la console — os.Stdout est aussi un io.Writer
ecrireDonnees(os.Stdout, "Bonjour depuis la console")

// Écrire dans un buffer mémoire — pour les tests
var buf bytes.Buffer
ecrireDonnees(&buf, "Bonjour depuis la mémoire")
```

> 💡 **C'est ça la vraie puissance** — `ecrireDonnees` ne sait pas si elle écrit dans un fichier, la console, ou la mémoire. Elle s'en fiche. Elle parle à une interface. C'est votre code qui décide de la destination. Ce niveau de flexibilité, atteint avec une interface de deux lignes, est ce qui rend Go si adaptable.

---

## 🛠️ Projet fil rouge — `gowatch` avec une architecture propre

On restructure `gowatch` avec des Structs bien définies, des méthodes, et une interface pour rendre le code extensible.

```go
package main

import (
    "encoding/json"
    "fmt"
    "io"
    "os"
    "runtime"
    "time"
)

// --- Types de données ---

// SystemeInfo contient les informations statiques de la machine
type SystemeInfo struct {
    OS           string `json:"os"`
    Architecture string `json:"architecture"`
    CPUCount     int    `json:"cpu_count"`
    GoVersion    string `json:"go_version"`
}

// Metrique représente une mesure système à un instant donné
type Metrique struct {
    Nom    string  `json:"nom"`
    Valeur float64 `json:"valeur"`
    Unite  string  `json:"unite"`
}

// String implémente fmt.Stringer pour un affichage propre
func (m Metrique) String() string {
    return fmt.Sprintf("%-15s : %.2f %s", m.Nom, m.Valeur, m.Unite)
}

// Snapshot regroupe toutes les données d'une collecte
type Snapshot struct {
    Timestamp time.Time   `json:"timestamp"`
    Systeme   SystemeInfo `json:"systeme"`
    Metriques []Metrique  `json:"metriques"`
}

// --- Interface de rendu ---

// Renderer définit comment afficher un Snapshot
// N'importe quel type avec une méthode Render implémente cette interface
type Renderer interface {
    Render(w io.Writer, snap Snapshot) error
}

// --- Implémentations du Renderer ---

// TextRenderer affiche en format texte lisible
type TextRenderer struct{}

func (r TextRenderer) Render(w io.Writer, snap Snapshot) error {
    fmt.Fprintf(w, "=== gowatch v0.4 — %s ===\n\n", snap.Timestamp.Format("15:04:05"))
    fmt.Fprintf(w, "Système  : %s/%s\n", snap.Systeme.OS, snap.Systeme.Architecture)
    fmt.Fprintf(w, "CPU(s)   : %d\n", snap.Systeme.CPUCount)
    fmt.Fprintf(w, "Go       : %s\n\n", snap.Systeme.GoVersion)

    for _, m := range snap.Metriques {
        fmt.Fprintln(w, m)  // Utilise String() automatiquement
    }
    return nil
}

// JSONRenderer affiche en format JSON indenté
type JSONRenderer struct{}

func (r JSONRenderer) Render(w io.Writer, snap Snapshot) error {
    donnees, err := json.MarshalIndent(snap, "", "  ")
    if err != nil {
        return fmt.Errorf("JSONRenderer.Render : %w", err)
    }
    _, err = fmt.Fprintln(w, string(donnees))
    return err
}

// --- Logique métier ---

// collecterSnapshot rassemble toutes les données système
func collecterSnapshot() Snapshot {
    return Snapshot{
        Timestamp: time.Now(),
        Systeme: SystemeInfo{
            OS:           runtime.GOOS,
            Architecture: runtime.GOARCH,
            CPUCount:     runtime.NumCPU(),
            GoVersion:    runtime.Version(),
        },
        Metriques: []Metrique{
            {Nom: "Goroutines", Valeur: float64(runtime.NumGoroutine()), Unite: "actives"},
        },
    }
}

// choisirRenderer retourne le bon Renderer selon le format demandé
func choisirRenderer(format string) Renderer {
    switch format {
    case "json":
        return JSONRenderer{}
    default:
        return TextRenderer{}
    }
}

// --- Point d'entrée ---

func main() {
    format := "texte"
    for _, arg := range os.Args[1:] {
        if arg == "--json" {
            format = "json"
        }
    }

    snap := collecterSnapshot()
    renderer := choisirRenderer(format)

    if err := renderer.Render(os.Stdout, snap); err != nil {
        fmt.Fprintf(os.Stderr, "Erreur de rendu : %v\n", err)
        os.Exit(1)
    }
}
```

**Ce que cette version apporte :**
- Des Structs claires avec des tags JSON (`json:"nom"`) pour la sérialisation
- Une interface `Renderer` — ajouter un nouveau format (CSV, HTML) ne changera rien au reste du code
- L'utilisation de `io.Writer` — la sortie peut être redirigée vers n'importe quel flux
- `fmt.Stringer` implémenté sur `Metrique` — l'affichage est propre automatiquement

```bash
go run main.go
go run main.go --json
```

---

## Ce qu'il faut retenir

1. **Go n'a pas de classes ni d'héritage** — il a des Structs et de la composition. Ce n'est pas un manque : c'est un design qui favorise la flexibilité et évite le couplage rigide des hiérarchies profondes.

2. **Les interfaces sont implicites** — si votre type a les bonnes méthodes, il implémente l'interface. Pas de `implements`, pas de déclaration. C'est le Duck Typing avec la sécurité du typage statique.

3. **La majuscule = exporté, la minuscule = privé** — c'est le seul système de visibilité en Go. Pas de `public`, `protected`, `private`. Simple, visible, universel.

4. **Récepteur par valeur = lecture, récepteur par pointeur = modification** — cette règle couvre 95% des cas. En cas de doute sur la performance, prenez le pointeur.

5. **`io.Writer` et `io.Reader` sont les interfaces les plus importantes** — dès que votre code parle à ces interfaces, il devient compatible avec des fichiers, la console, des connexions réseau, des buffers mémoire. Gratuitement.

---

## Pour aller plus loin

- 📄 [Effective Go — Interfaces](https://go.dev/doc/effective_go#interfaces) — La référence officielle sur les interfaces Go
- 📄 [Go Data Structures](https://research.swtch.com/godata) — Comment Go représente les Structs en mémoire
- 📄 [Composition over inheritance in Go](https://go.dev/doc/faq#Is_Go_an_object-oriented_language) — La FAQ officielle sur la POO en Go

---

<div align="center">

[⬅️ Chapitre 2.2 — La logique sans fioritures](./02-logique-sans-fioritures.md) · [👉 Module 03 — Le Super-Pouvoir](../module-03-super-pouvoir/README.md)

</div>
