# Chapitre 4.2 — Tooling et Qualité Production

> *"Un code sans test est un code dont on suppose qu'il fonctionne.
> Un code avec des tests est un code dont on sait qu'il fonctionne."*

---

## Le problème

Votre API fonctionne. Vous l'avez testée manuellement avec `curl`. Les réponses sont bonnes.

Mais dans trois semaines, vous modifiez le handler `handleMetrics` pour ajouter un champ. Et sans le savoir, vous cassez le format JSON que `handleHistory` attendait. Vous le découvrez en production, à 23h, quand un utilisateur signale que son dashboard est vide.

Ce scénario se produit dans chaque projet qui n'a pas de tests automatisés. La question n'est pas *si* ça arrivera — c'est *quand*.

Go a une réponse à ce problème. Et comme tout ce que fait Go, elle est simple, intégrée, et sans dépendances externes.

---

## L'intuition

### Les tests en Go ne sont pas une corvée

Dans beaucoup d'écosystèmes, écrire des tests demande d'installer un framework de test, d'apprendre sa syntaxe, de configurer un runner, et de gérer des assertions complexes.

En Go, `go test ./...` est la commande la plus puissante de votre arsenal. Elle est intégrée, rapide, et fonctionne sans rien installer. Le package `testing` de la bibliothèque standard couvre 95% des besoins.

La philosophie est la même qu'ailleurs en Go : **moins de magie, plus de contrôle**.

### Le Table-Driven Test — L'idiome central

L'approche la plus caractéristique des tests Go est le **Table-Driven Test** : vous définissez un tableau de cas de test, et vous les faites tous tourner dans une boucle. Un seul test couvre des dizaines de scénarios.

---

## La solution Go

### Les bases — Un premier test

Par convention, les fichiers de test en Go se terminent par `_test.go`. Go les compile et les exécute uniquement lors d'un `go test` — jamais dans le binaire final.

```go
// store/memory_test.go
package store

import (
    "testing"
    "time"
)

func TestStore_Ajouter(t *testing.T) {
    s := New()

    snap := Snapshot{
        Timestamp: time.Now(),
        Metriques: []Metrique{
            {Source: "CPU", Valeur: 23.4, Unite: "%"},
        },
    }

    s.Ajouter(snap)

    dernier, ok := s.Dernier()
    if !ok {
        t.Fatal("Dernier() devrait retourner true après un ajout")
    }

    if dernier.Metriques[0].Source != "CPU" {
        t.Errorf("Source attendue 'CPU', obtenu '%s'", dernier.Metriques[0].Source)
    }
}
```

**Lancer les tests :**
```bash
go test ./...
# ok  github.com/votre-pseudo/gohub/store  0.002s

go test -v ./...  # Mode verbeux — affiche chaque test
# === RUN   TestStore_Ajouter
# --- PASS: TestStore_Ajouter (0.00s)
# ok  github.com/votre-pseudo/gohub/store  0.002s
```

> 💡 **Les fonctions de test** — Elles commencent toujours par `Test`, prennent `*testing.T` en paramètre, et ne retournent rien. Le framework détecte automatiquement toutes les fonctions qui respectent cette signature.

**Les méthodes essentielles de `*testing.T` :**

```go
t.Error("message")      // Marque le test en échec, continue l'exécution
t.Errorf("format", ...) // Idem avec formatting
t.Fatal("message")      // Marque le test en échec, arrête immédiatement
t.Fatalf("format", ...) // Idem avec formatting
t.Log("message")        // Log visible uniquement avec -v
t.Helper()              // Marque la fonction comme helper (meilleure trace d'erreur)
```

---

### Table-Driven Tests — L'idiome Go par excellence

Au lieu d'écrire un test par scénario, on définit une table de cas et on les itère :

```go
// store/memory_test.go

func TestStore_Historique(t *testing.T) {
    // Définition de la table de cas
    tests := []struct {
        nom             string // Nom du cas — pour identifier l'échec
        nbSnapshots     int    // Nombre de snapshots à insérer
        limite          int    // Limite demandée
        attendu         int    // Nombre de résultats attendus
    }{
        {
            nom:         "limite inférieure au total",
            nbSnapshots: 10,
            limite:      5,
            attendu:     5,
        },
        {
            nom:         "limite supérieure au total",
            nbSnapshots: 3,
            limite:      10,
            attendu:     3,
        },
        {
            nom:         "store vide",
            nbSnapshots: 0,
            limite:      5,
            attendu:     0,
        },
        {
            nom:         "limite zéro",
            nbSnapshots: 5,
            limite:      0,
            attendu:     0,
        },
    }

    // Itération sur tous les cas
    for _, tc := range tests {
        tc := tc // Capture pour les sous-tests parallèles

        // t.Run crée un sous-test nommé — chaque cas est indépendant
        t.Run(tc.nom, func(t *testing.T) {
            s := New()

            for i := 0; i < tc.nbSnapshots; i++ {
                s.Ajouter(Snapshot{Timestamp: time.Now()})
            }

            résultat := s.Historique(tc.limite)

            if len(résultat) != tc.attendu {
                t.Errorf("Historique(%d) avec %d snapshots : attendu %d, obtenu %d",
                    tc.limite, tc.nbSnapshots, tc.attendu, len(résultat))
            }
        })
    }
}
```

**Lancer un cas spécifique :**
```bash
go test -run TestStore_Historique/limite_inférieure_au_total ./store/
```

> 💡 **Pourquoi le Table-Driven Test ?**
> - Ajouter un nouveau cas = ajouter une ligne dans la table
> - Quand un test échoue, le nom du cas (`tc.nom`) identifie immédiatement le problème
> - Tous les cas partagent la même logique de vérification — pas de duplication
> - C'est la convention Go : quand vous lisez du code Go open source, vous verrez ce pattern partout

---

### Tester les handlers HTTP — `httptest`

Go fournit le package `net/http/httptest` pour tester les handlers HTTP sans démarrer un vrai serveur.

```go
// api/server_test.go
package api

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/votre-pseudo/gohub/store"
)

// setupTestServer crée un serveur de test avec des données initiales
func setupTestServer(t *testing.T) *Server {
    t.Helper() // Marque cette fonction comme helper

    s := store.New()
    s.Ajouter(store.Snapshot{
        Timestamp: time.Now(),
        Metriques: []store.Metrique{
            {Source: "CPU", Valeur: 23.4, Unite: "%"},
        },
    })

    return New(s, "test")
}

func TestHandleHealth(t *testing.T) {
    srv := setupTestServer(t)

    // Créer une requête HTTP de test
    req := httptest.NewRequest(http.MethodGet, "/health", nil)

    // Créer un ResponseRecorder — enregistre la réponse sans réseau
    rec := httptest.NewRecorder()

    // Appeler le handler directement
    srv.handleHealth(rec, req)

    // Vérifier le code de statut
    if rec.Code != http.StatusOK {
        t.Errorf("Code attendu %d, obtenu %d", http.StatusOK, rec.Code)
    }

    // Vérifier le Content-Type
    contentType := rec.Header().Get("Content-Type")
    if contentType != "application/json" {
        t.Errorf("Content-Type attendu 'application/json', obtenu '%s'", contentType)
    }

    // Vérifier le body JSON
    var reponse map[string]string
    if err := json.NewDecoder(rec.Body).Decode(&reponse); err != nil {
        t.Fatalf("Impossible de décoder la réponse JSON : %v", err)
    }

    if reponse["status"] != "ok" {
        t.Errorf("Status attendu 'ok', obtenu '%s'", reponse["status"])
    }
}

func TestHandleMetrics_TableDriven(t *testing.T) {
    tests := []struct {
        nom            string
        methode        string
        storeVide      bool
        statusAttendu  int
    }{
        {
            nom:           "GET avec données",
            methode:       http.MethodGet,
            storeVide:     false,
            statusAttendu: http.StatusOK,
        },
        {
            nom:           "GET store vide",
            methode:       http.MethodGet,
            storeVide:     true,
            statusAttendu: http.StatusNotFound,
        },
        {
            nom:           "POST non autorisé",
            methode:       http.MethodPost,
            storeVide:     false,
            statusAttendu: http.StatusMethodNotAllowed,
        },
        {
            nom:           "DELETE non autorisé",
            methode:       http.MethodDelete,
            storeVide:     false,
            statusAttendu: http.StatusMethodNotAllowed,
        },
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.nom, func(t *testing.T) {
            var srv *Server
            if tc.storeVide {
                srv = New(store.New(), "test")
            } else {
                srv = setupTestServer(t)
            }

            req := httptest.NewRequest(tc.methode, "/api/metrics", nil)
            rec := httptest.NewRecorder()

            srv.handleMetrics(rec, req)

            if rec.Code != tc.statusAttendu {
                t.Errorf("Code attendu %d, obtenu %d", tc.statusAttendu, rec.Code)
            }
        })
    }
}
```

**Lancer avec couverture de code :**
```bash
go test -cover ./...
# ok  github.com/votre-pseudo/gohub/store  coverage: 87.5% of statements
# ok  github.com/votre-pseudo/gohub/api    coverage: 72.3% of statements

# Rapport de couverture visuel dans le navigateur
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

### Mocking avec les interfaces — Tester sans dépendances

Le vrai pouvoir des interfaces Go brille dans les tests. Quand votre code dépend d'une interface plutôt que d'une implémentation concrète, vous pouvez substituer un mock en test.

```go
// Définir l'interface du store
type Storer interface {
    Ajouter(snap store.Snapshot)
    Dernier() (store.Snapshot, bool)
    Historique(limite int) []store.Snapshot
}

// MockStore — implémentation de test
type MockStore struct {
    snapshots []store.Snapshot
    // Champs pour vérifier les appels
    AjouterAppelé  bool
    DernierAppelé  bool
}

func (m *MockStore) Ajouter(snap store.Snapshot) {
    m.AjouterAppelé = true
    m.snapshots = append(m.snapshots, snap)
}

func (m *MockStore) Dernier() (store.Snapshot, bool) {
    m.DernierAppelé = true
    if len(m.snapshots) == 0 {
        return store.Snapshot{}, false
    }
    return m.snapshots[len(m.snapshots)-1], true
}

func (m *MockStore) Historique(limite int) []store.Snapshot {
    if limite > len(m.snapshots) {
        return m.snapshots
    }
    return m.snapshots[len(m.snapshots)-limite:]
}

// Test avec mock — aucune dépendance sur le vrai Store
func TestHandleIngest_AvecMock(t *testing.T) {
    mock := &MockStore{}
    srv := &Server{store: mock, startTime: time.Now(), version: "test", mux: http.NewServeMux()}

    body := strings.NewReader(`{"metriques":[{"source":"test","valeur":42,"unite":"ms"}]}`)
    req := httptest.NewRequest(http.MethodPost, "/api/metrics/ingest", body)
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    srv.handleIngest(rec, req)

    if rec.Code != http.StatusCreated {
        t.Errorf("Status attendu %d, obtenu %d", http.StatusCreated, rec.Code)
    }

    if !mock.AjouterAppelé {
        t.Error("Ajouter() aurait dû être appelé")
    }
}
```

> 💡 **Interface dans le Server** — Pour que le mock fonctionne, `Server` doit dépendre d'une interface `Storer`, pas du type concret `*store.Store`. C'est le principe de l'**injection de dépendances** — vous injectez ce dont le code a besoin, pas ce qu'il crée lui-même.

---

### Benchmarks — Mesurer pour ne pas supposer

Go intègre un système de benchmarking dans le même package `testing`. Les fonctions de benchmark commencent par `Bench` :

```go
// store/memory_test.go

func BenchmarkStore_Ajouter(b *testing.B) {
    s := New()
    snap := Snapshot{
        Timestamp: time.Now(),
        Metriques: []Metrique{
            {Source: "CPU", Valeur: 23.4, Unite: "%"},
            {Source: "RAM", Valeur: 67.1, Unite: "%"},
        },
    }

    // b.N est automatiquement ajusté par Go pour obtenir des mesures stables
    for i := 0; i < b.N; i++ {
        s.Ajouter(snap)
    }
}

func BenchmarkStore_Historique(b *testing.B) {
    s := New()

    // Setup — ne pas mesurer cette partie
    for i := 0; i < 1000; i++ {
        s.Ajouter(Snapshot{Timestamp: time.Now()})
    }

    b.ResetTimer() // Réinitialiser le chrono après le setup

    for i := 0; i < b.N; i++ {
        s.Historique(100)
    }
}
```

**Lancer les benchmarks :**
```bash
go test -bench=. ./store/
# BenchmarkStore_Ajouter-8       5823901    205.3 ns/op
# BenchmarkStore_Historique-8    2156789    556.8 ns/op

# Avec la mémoire allouée
go test -bench=. -benchmem ./store/
# BenchmarkStore_Ajouter-8    5823901    205.3 ns/op    256 B/op    2 allocs/op
```

**Lire les résultats :**
- `5823901` — nombre d'itérations exécutées
- `205.3 ns/op` — temps moyen par opération
- `256 B/op` — mémoire allouée par opération
- `2 allocs/op` — nombre d'allocations mémoire par opération

> 💡 **Pourquoi les benchmarks ?** — La réponse intuitive à "mon code est lent" est souvent fausse. Les benchmarks vous montrent *où* c'est vraiment lent, et *combien* une optimisation améliore les choses. Optimiser sans mesurer, c'est jouer à la loterie.

---

### Tests parallèles — Accélérer la suite de tests

Pour les tests qui peuvent s'exécuter en parallèle sans interférer :

```go
func TestStore_Concurrent(t *testing.T) {
    t.Parallel() // Ce test peut tourner en parallèle avec d'autres

    s := New()
    var wg sync.WaitGroup

    // 100 goroutines ajoutent en même temps — test de race condition
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            s.Ajouter(Snapshot{Timestamp: time.Now()})
        }()
    }

    wg.Wait()

    historique := s.Historique(1000)
    if len(historique) != 100 {
        t.Errorf("Attendu 100 snapshots, obtenu %d", len(historique))
    }
}
```

**Lancer avec le race detector :**
```bash
go test -race ./...
# Détecte les race conditions dans vos tests — indispensable
```

---

### Go Modules — Gérer les dépendances proprement

Go Modules est le système de gestion de dépendances intégré depuis Go 1.11. Deux fichiers le constituent :

- **`go.mod`** — liste vos dépendances directes et la version de Go
- **`go.sum`** — les checksums cryptographiques de chaque dépendance (ne pas modifier à la main)

```bash
# Ajouter une dépendance
go get github.com/go-chi/chi/v5@latest

# Mettre à jour toutes les dépendances
go get -u ./...

# Nettoyer les dépendances inutilisées
go mod tidy

# Télécharger les dépendances localement (pour le mode offline)
go mod download

# Vendoriser les dépendances (copie locale dans /vendor)
go mod vendor
```

**`go.mod` typique après ajout de dépendances :**
```
module github.com/votre-pseudo/gohub

go 1.23

require (
    github.com/go-chi/chi/v5 v5.0.10
    github.com/mattn/go-sqlite3 v1.14.18
)
```

> ⚠️ **Commitez toujours `go.sum`** — Ce fichier garantit la reproductibilité des builds. Sans lui, deux développeurs peuvent obtenir des versions différentes d'une même dépendance.

> 💡 **`go mod tidy`** — Lancez cette commande avant chaque commit. Elle ajoute les dépendances manquantes et supprime celles qui ne sont plus utilisées. C'est le `npm prune` + `npm install` de Go, en une seule commande.

---

## 🛠️ Projet fil rouge — Tests complets de `gohub`

On ajoute une suite de tests complète au projet `gohub`.

**Créez `store/memory_test.go` :**

```go
package store

import (
    "sync"
    "testing"
    "time"
)

func TestStore_VideInitialement(t *testing.T) {
    s := New()
    _, ok := s.Dernier()
    if ok {
        t.Error("Un store vide ne devrait pas avoir de dernier élément")
    }
}

func TestStore_AjouterEtDernier(t *testing.T) {
    s := New()
    snap := Snapshot{
        Timestamp: time.Now(),
        Metriques: []Metrique{{Source: "CPU", Valeur: 42.0, Unite: "%"}},
    }

    s.Ajouter(snap)

    dernier, ok := s.Dernier()
    if !ok {
        t.Fatal("Dernier() devrait retourner true")
    }
    if dernier.ID != 1 {
        t.Errorf("Premier ID attendu 1, obtenu %d", dernier.ID)
    }
    if dernier.Metriques[0].Source != "CPU" {
        t.Errorf("Source attendue 'CPU', obtenu '%s'", dernier.Metriques[0].Source)
    }
}

func TestStore_Historique(t *testing.T) {
    tests := []struct {
        nom         string
        inserts     int
        limite      int
        attendu     int
    }{
        {"limite < total",    10, 5,  5},
        {"limite > total",    3,  10, 3},
        {"store vide",        0,  5,  0},
        {"limite zéro",       5,  0,  0},
        {"limite = total",    5,  5,  5},
    }

    for _, tc := range tests {
        tc := tc
        t.Run(tc.nom, func(t *testing.T) {
            s := New()
            for i := 0; i < tc.inserts; i++ {
                s.Ajouter(Snapshot{Timestamp: time.Now()})
            }
            if got := len(s.Historique(tc.limite)); got != tc.attendu {
                t.Errorf("Historique(%d) avec %d inserts : attendu %d, obtenu %d",
                    tc.limite, tc.inserts, tc.attendu, got)
            }
        })
    }
}

func TestStore_LimiteMemoireMaximale(t *testing.T) {
    s := New()
    // Insérer plus de 1000 snapshots
    for i := 0; i < 1100; i++ {
        s.Ajouter(Snapshot{Timestamp: time.Now()})
    }
    // Le store ne doit pas dépasser 1000
    if got := len(s.Historique(2000)); got > 1000 {
        t.Errorf("Le store devrait être limité à 1000, contient %d", got)
    }
}

func TestStore_ConcurrenceSafe(t *testing.T) {
    t.Parallel()
    s := New()
    var wg sync.WaitGroup
    const n = 500

    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            s.Ajouter(Snapshot{Timestamp: time.Now()})
        }()
    }
    wg.Wait()

    if got := len(s.Historique(n)); got != n {
        t.Errorf("Attendu %d snapshots, obtenu %d", n, got)
    }
}

func BenchmarkStore_Ajouter(b *testing.B) {
    s := New()
    snap := Snapshot{
        Timestamp: time.Now(),
        Metriques: []Metrique{
            {Source: "CPU", Valeur: 23.4, Unite: "%"},
            {Source: "RAM", Valeur: 67.1, Unite: "%"},
        },
    }
    for i := 0; i < b.N; i++ {
        s.Ajouter(snap)
    }
}

func BenchmarkStore_Historique100(b *testing.B) {
    s := New()
    for i := 0; i < 1000; i++ {
        s.Ajouter(Snapshot{Timestamp: time.Now()})
    }
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        s.Historique(100)
    }
}
```

**Lancez tout :**
```bash
# Tests unitaires
go test ./...

# Tests avec race detector
go test -race ./...

# Tests avec couverture
go test -cover ./...

# Benchmarks
go test -bench=. -benchmem ./store/

# Rapport de couverture HTML
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Résultat attendu :**
```
=== RUN   TestStore_VideInitialement
--- PASS: TestStore_VideInitialement (0.00s)
=== RUN   TestStore_AjouterEtDernier
--- PASS: TestStore_AjouterEtDernier (0.00s)
=== RUN   TestStore_Historique
=== RUN   TestStore_Historique/limite_<_total
=== RUN   TestStore_Historique/limite_>_total
=== RUN   TestStore_Historique/store_vide
=== RUN   TestStore_Historique/limite_zéro
=== RUN   TestStore_Historique/limite_=_total
--- PASS: TestStore_Historique (0.00s)
=== RUN   TestStore_LimiteMemoireMaximale
--- PASS: TestStore_LimiteMemoireMaximale (0.00s)
=== RUN   TestStore_ConcurrenceSafe
--- PASS: TestStore_ConcurrenceSafe (0.01s)
ok  github.com/votre-pseudo/gohub/store  coverage: 100.0% of statements
```

100% de couverture sur le store. Et si vous modifiez la logique de `Historique` demain, les tests vous le diront immédiatement.

---

## Ce qu'il faut retenir

1. **`go test ./...`** — La commande la plus importante après `go build`. Intégrée, sans framework, sans configuration. Lancez-la avant chaque commit.

2. **Table-Driven Tests** — L'idiome central des tests Go. Un tableau de cas, une boucle, des sous-tests nommés avec `t.Run`. Ajouter un cas = ajouter une ligne.

3. **`httptest.NewRequest` + `httptest.NewRecorder`** — Tester les handlers HTTP sans démarrer un vrai serveur. Rapide, isolé, déterministe.

4. **Les interfaces permettent le mocking** — Si votre code dépend d'une interface, vous pouvez substituer n'importe quelle implémentation en test. C'est l'injection de dépendances à la Go.

5. **`go test -race`** — Détecte les race conditions dans vos tests. À activer systématiquement en CI. La couverture ne détecte pas les bugs de concurrence — le race detector, si.

6. **`go mod tidy`** — Avant chaque commit. Commitez `go.sum`. Ne modifiez jamais `go.sum` à la main.

---

## Pour aller plus loin

- 📄 [Testing — Documentation officielle](https://pkg.go.dev/testing)
- 📄 [Table Driven Tests — Blog officiel Go](https://go.dev/wiki/TableDrivenTests)
- 📄 [Using Go Modules — Blog officiel Go](https://go.dev/blog/using-go-modules)
- 🔧 [testify](https://github.com/stretchr/testify) — Assertions plus expressives si `if got != want` devient verbeux

---

<div align="center">

[⬅️ Chapitre 4.1 — Cloud Native et Microservices](./01-cloud-native-microservices.md) · [👉 Chapitre 4.3 — Déploiement et DevOps](./03-deploiement-devops.md)

</div>
