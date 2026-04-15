# Chapitre 5.1 — Réseau et Sécurité Système

> *"Le réseau est la frontière entre votre code et le monde réel.
> Go vous donne les outils pour la traverser proprement — dans les deux sens."*

---

## Le problème

Docker est écrit en Go. Kubernetes est écrit en Go. Terraform est écrit en Go. Cloudflare fait tourner ses proxies réseau en Go.

Ce n'est pas un hasard.

Go est particulièrement adapté au code réseau pour trois raisons concrètes : ses goroutines gèrent des milliers de connexions simultanées sans douleur, sa bibliothèque standard `net` couvre TCP, UDP, DNS et HTTP sans dépendance externe, et ses primitives cryptographiques sont solides, bien testées, et idiomatiques.

Ce chapitre vous emmène du serveur HTTP qu'on a construit au Module 04 vers le réseau bas niveau : connexions TCP brutes, scanners de ports, tunneling, et cryptographie.

---

## L'intuition

### TCP, c'est quoi vraiment ?

HTTP, que vous connaissez bien maintenant, est construit **au-dessus** de TCP. TCP est le protocole de transport — il garantit que les données arrivent dans l'ordre et sans erreur. HTTP dit simplement *ce qu'on envoie* dans ces données.

Quand vous écrivez du code TCP bas niveau, vous avez accès à la couche en dessous de HTTP. C'est là que vivent les protocoles custom, les outils de monitoring réseau, les scanners, et les tunnels.

En Go, une connexion TCP s'utilise comme n'importe quel `io.Reader` / `io.Writer` — la même interface qu'un fichier ou un buffer mémoire. C'est la cohérence du design Go en action.

---

## La solution Go

### TCP bas niveau — Serveur et client

**Serveur TCP simple :**
```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "strings"
)

func gererConnexion(conn net.Conn) {
    defer conn.Close()

    adresse := conn.RemoteAddr().String()
    fmt.Printf("Nouvelle connexion depuis %s\n", adresse)

    scanner := bufio.NewScanner(conn)
    for scanner.Scan() {
        ligne := scanner.Text()
        fmt.Printf("[%s] reçu : %s\n", adresse, ligne)

        // Répondre en majuscules — exemple trivial
        reponse := strings.ToUpper(ligne) + "\n"
        conn.Write([]byte(reponse))
    }

    fmt.Printf("Connexion fermée : %s\n", adresse)
}

func main() {
    listener, err := net.Listen("tcp", ":9000")
    if err != nil {
        panic(err)
    }
    defer listener.Close()

    fmt.Println("Serveur TCP en écoute sur :9000")

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Erreur Accept :", err)
            continue
        }
        // Chaque connexion dans sa goroutine — concurrent par nature
        go gererConnexion(conn)
    }
}
```

**Client TCP simple :**
```go
func main() {
    conn, err := net.Dial("tcp", "localhost:9000")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Envoyer un message
    fmt.Fprintln(conn, "bonjour depuis le client")

    // Lire la réponse
    reponse, err := bufio.NewReader(conn).ReadString('\n')
    if err != nil {
        panic(err)
    }
    fmt.Printf("Réponse du serveur : %s", reponse)
    // Réponse du serveur : BONJOUR DEPUIS LE CLIENT
}
```

> 💡 **`net.Dial` et `net.Listen`** — Ces deux fonctions sont le point d'entrée de tout code réseau Go. `Listen` crée un socket serveur. `Dial` crée une connexion client. Les deux retournent une interface `net.Conn` qui est à la fois `io.Reader` et `io.Writer` — vous pouvez y appliquer tous les outils que vous connaissez déjà.

---

### Timeouts sur les connexions réseau

Le réseau est imprévisible. Sans timeout, une connexion bloquée peut geler votre programme indéfiniment.

```go
import "time"

// Dial avec timeout
conn, err := net.DialTimeout("tcp", "localhost:9000", 3*time.Second)
if err != nil {
    // Erreur de connexion ou timeout
    return fmt.Errorf("connexion impossible : %w", err)
}

// Timeout de lecture — si aucune donnée n'arrive dans 5s
conn.SetReadDeadline(time.Now().Add(5 * time.Second))

// Timeout d'écriture
conn.SetWriteDeadline(time.Now().Add(5 * time.Second))

// Timeout global sur toutes les opérations
conn.SetDeadline(time.Now().Add(30 * time.Second))
```

---

### Scanner de ports concurrent — Le cas d'usage parfait

Un scanner de ports teste si des ports TCP sont ouverts sur une machine. C'est le cas d'usage idéal pour les goroutines : chaque port est testé indépendamment, en parallèle, et on agrège les résultats.

```go
package main

import (
    "context"
    "fmt"
    "net"
    "os"
    "sort"
    "sync"
    "time"
)

type ResultatPort struct {
    Port   int
    Ouvert bool
    Erreur error
}

// testerPort tente une connexion TCP sur le port donné
func testerPort(ctx context.Context, hote string, port int) ResultatPort {
    adresse := fmt.Sprintf("%s:%d", hote, port)

    // Utiliser DialContext pour respecter le contexte d'annulation
    var d net.Dialer
    conn, err := d.DialContext(ctx, "tcp", adresse)

    if err != nil {
        return ResultatPort{Port: port, Ouvert: false}
    }
    conn.Close()
    return ResultatPort{Port: port, Ouvert: true}
}

// scannerPorts scanne une plage de ports de façon concurrente
func scannerPorts(ctx context.Context, hote string, portDebut, portFin, concurrence int) []ResultatPort {
    resultats := make(chan ResultatPort, portFin-portDebut+1)
    semaphore := make(chan struct{}, concurrence) // Limiter la concurrence

    var wg sync.WaitGroup

    for port := portDebut; port <= portFin; port++ {
        // Vérifier si le contexte est annulé avant de lancer
        select {
        case <-ctx.Done():
            break
        default:
        }

        wg.Add(1)
        p := port
        go func() {
            defer wg.Done()

            semaphore <- struct{}{}        // Acquérir le slot
            defer func() { <-semaphore }() // Libérer le slot

            // Timeout court par port — on ne veut pas attendre trop longtemps
            portCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
            defer cancel()

            resultats <- testerPort(portCtx, hote, p)
        }()
    }

    // Fermer le channel quand tout est terminé
    go func() {
        wg.Wait()
        close(resultats)
    }()

    // Collecter les résultats
    var tous []ResultatPort
    for r := range resultats {
        tous = append(tous, r)
    }

    // Trier par numéro de port pour l'affichage
    sort.Slice(tous, func(i, j int) bool {
        return tous[i].Port < tous[j].Port
    })

    return tous
}

// nomService retourne le nom conventionnel d'un port connu
func nomService(port int) string {
    services := map[int]string{
        21: "FTP", 22: "SSH", 23: "Telnet", 25: "SMTP",
        53: "DNS", 80: "HTTP", 110: "POP3", 143: "IMAP",
        443: "HTTPS", 465: "SMTPS", 587: "SMTP/TLS",
        993: "IMAPS", 995: "POP3S", 3306: "MySQL",
        5432: "PostgreSQL", 6379: "Redis", 8080: "HTTP-Alt",
        8443: "HTTPS-Alt", 27017: "MongoDB",
    }
    if nom, ok := services[port]; ok {
        return nom
    }
    return "?"
}

func main() {
    hote := "localhost"
    portDebut := 1
    portFin := 1024
    concurrence := 100 // 100 connexions simultanées max

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    fmt.Printf("Scan de %s (ports %d-%d, concurrence=%d)\n\n",
        hote, portDebut, portFin, concurrence)

    debut := time.Now()
    resultats := scannerPorts(ctx, hote, portDebut, portFin, concurrence)
    duree := time.Since(debut)

    // Afficher uniquement les ports ouverts
    ouverts := 0
    for _, r := range resultats {
        if r.Ouvert {
            ouverts++
            fmt.Printf("  ✓ Port %-6d : ouvert (%s)\n", r.Port, nomService(r.Port))
        }
    }

    fmt.Printf("\n%d ports ouverts sur %d testés — terminé en %v\n",
        ouverts, portFin-portDebut+1, duree.Round(time.Millisecond))
}
```

**Résultat :**
```
Scan de localhost (ports 1-1024, concurrence=100)

  ✓ Port 22     : ouvert (SSH)
  ✓ Port 80     : ouvert (HTTP)
  ✓ Port 443    : ouvert (HTTPS)
  ✓ Port 8080   : ouvert (HTTP-Alt)

4 ports ouverts sur 1024 testés — terminé en 1.2s
```

> 💡 **Le sémaphore `chan struct{}`** — On limite la concurrence à 100 connexions simultanées avec un channel buffered utilisé comme sémaphore. Trop de connexions simultanées saturerait les ressources système ou déclencherait des protections anti-flood. C'est le pattern de rate-limiting le plus simple en Go.

---

### Tunneling TCP — Rediriger le trafic

Un tunnel TCP reçoit des connexions sur un port local et les redirige vers une destination distante. C'est la base de nombreux outils réseau : proxies, VPNs simples, forwarding de ports.

```go
package main

import (
    "fmt"
    "io"
    "net"
)

// tunnel copie les données dans les deux sens entre deux connexions
func tunnel(local, distant net.Conn) {
    defer local.Close()
    defer distant.Close()

    // Copier dans les deux sens simultanément
    // io.Copy bloque jusqu'à la fin ou une erreur
    done := make(chan struct{}, 2)

    go func() {
        io.Copy(distant, local) // local → distant
        done <- struct{}{}
    }()

    go func() {
        io.Copy(local, distant) // distant → local
        done <- struct{}{}
    }()

    // Attendre que l'une des deux directions se ferme
    <-done
}

// demarrerTunnel crée un tunnel TCP entre localAddr et distantAddr
func demarrerTunnel(localAddr, distantAddr string) error {
    listener, err := net.Listen("tcp", localAddr)
    if err != nil {
        return fmt.Errorf("impossible d'écouter sur %s : %w", localAddr, err)
    }
    defer listener.Close()

    fmt.Printf("Tunnel actif : %s → %s\n", localAddr, distantAddr)

    for {
        local, err := listener.Accept()
        if err != nil {
            return fmt.Errorf("erreur Accept : %w", err)
        }

        go func() {
            // Se connecter à la destination
            distant, err := net.Dial("tcp", distantAddr)
            if err != nil {
                fmt.Printf("Impossible de joindre %s : %v\n", distantAddr, err)
                local.Close()
                return
            }

            fmt.Printf("Connexion tunnelisée : %s → %s\n",
                local.RemoteAddr(), distantAddr)

            tunnel(local, distant)
        }()
    }
}

func main() {
    // Tunnel : écoute sur :9001, redirige vers gohub sur :8080
    if err := demarrerTunnel(":9001", "localhost:8080"); err != nil {
        fmt.Println("Erreur :", err)
    }
}
```

```bash
go run main.go
# Tunnel actif : :9001 → localhost:8080

# Dans un autre terminal
curl http://localhost:9001/health
# {"status":"ok",...}  ← La requête passe par le tunnel
```

> 💡 **`io.Copy`** — Cette fonction copie depuis un `io.Reader` vers un `io.Writer` jusqu'à EOF ou erreur. Elle est la brique de base de tout code réseau Go : `net.Conn` implémente les deux interfaces, donc `io.Copy` fonctionne directement sur n'importe quelle connexion TCP. Deux goroutines avec deux `io.Copy` dans des directions opposées — c'est tout ce qu'il faut pour un tunnel bidirectionnel.

---

### TLS — HTTPS et connexions sécurisées

Go intègre une implémentation TLS complète dans `crypto/tls`. Pas besoin d'OpenSSL.

**Générer un certificat auto-signé pour les tests :**
```bash
# Générer une clé privée et un certificat auto-signé
openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt \
    -days 365 -nodes -subj "/CN=localhost"
```

**Serveur HTTPS avec `crypto/tls` :**
```go
package main

import (
    "crypto/tls"
    "fmt"
    "net/http"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintln(w, "Connexion sécurisée !")
        fmt.Fprintf(w, "Protocol TLS : %v\n", r.TLS.Version)
    })

    // Configuration TLS — bonnes pratiques de sécurité
    config := &tls.Config{
        MinVersion: tls.VersionTLS12,          // TLS 1.2 minimum
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
        },
        PreferServerCipherSuites: true,
    }

    srv := &http.Server{
        Addr:      ":8443",
        Handler:   mux,
        TLSConfig: config,
    }

    fmt.Println("Serveur HTTPS sur :8443")
    // ListenAndServeTLS charge les certificats depuis les fichiers
    if err := srv.ListenAndServeTLS("server.crt", "server.key"); err != nil {
        panic(err)
    }
}
```

**Client HTTPS avec vérification de certificat personnalisée :**
```go
func creerClientHTTPS(certFile string) (*http.Client, error) {
    // Charger le certificat CA pour valider le serveur
    caCert, err := os.ReadFile(certFile)
    if err != nil {
        return nil, fmt.Errorf("lecture certificat CA : %w", err)
    }

    pool := x509.NewCertPool()
    if !pool.AppendCertsFromPEM(caCert) {
        return nil, fmt.Errorf("impossible d'ajouter le certificat CA")
    }

    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            RootCAs:    pool,
            MinVersion: tls.VersionTLS12,
        },
    }

    return &http.Client{
        Transport: transport,
        Timeout:   10 * time.Second,
    }, nil
}
```

---

### Cryptographie — Les primitives essentielles

Go intègre toutes les primitives cryptographiques modernes dans `crypto/*`. Voici celles que vous utiliserez le plus souvent.

**Hashing — SHA-256 et bcrypt :**
```go
import (
    "crypto/sha256"
    "crypto/hmac"
    "encoding/hex"
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

// Hash SHA-256 — pour l'intégrité des données (pas les mots de passe !)
func hashSHA256(data string) string {
    h := sha256.Sum256([]byte(data))
    return hex.EncodeToString(h[:])
}

// HMAC-SHA256 — hash avec clé secrète, pour les signatures
func hmacSHA256(data, cle string) string {
    h := hmac.New(sha256.New, []byte(cle))
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

// Vérifier un HMAC — résistant aux timing attacks
func verifierHMAC(data, cle, signature string) bool {
    expected := hmacSHA256(data, cle)
    // hmac.Equal compare de façon à éviter les timing attacks
    return hmac.Equal([]byte(signature), []byte(expected))
}

// bcrypt — pour les mots de passe (lent par design, résistant au brute-force)
func hasherMotDePasse(mdp string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(mdp), bcrypt.DefaultCost)
    return string(hash), err
}

func verifierMotDePasse(mdp, hash string) bool {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(mdp)) == nil
}

func main() {
    // SHA-256
    fmt.Println(hashSHA256("gohub"))
    // 3b4c...  (toujours le même pour la même entrée)

    // HMAC
    sig := hmacSHA256("données importantes", "clé-secrète")
    fmt.Println(verifierHMAC("données importantes", "clé-secrète", sig)) // true
    fmt.Println(verifierHMAC("données modifiées",   "clé-secrète", sig)) // false

    // bcrypt
    hash, _ := hasherMotDePasse("mon-mot-de-passe")
    fmt.Println(verifierMotDePasse("mon-mot-de-passe", hash))  // true
    fmt.Println(verifierMotDePasse("mauvais-mdp",      hash))  // false
}
```

**Chiffrement symétrique — AES-GCM :**
```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "io"
)

// ChiffrerAESGCM chiffre des données avec AES-256-GCM
// AES-GCM est authentifié — il détecte toute modification des données
func ChiffrerAESGCM(texte, cle []byte) ([]byte, error) {
    block, err := aes.NewCipher(cle) // cle doit faire 16, 24, ou 32 bytes
    if err != nil {
        return nil, fmt.Errorf("création cipher : %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("création GCM : %w", err)
    }

    // Nonce aléatoire — doit être unique pour chaque chiffrement avec la même clé
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, fmt.Errorf("génération nonce : %w", err)
    }

    // Chiffrer et préfixer avec le nonce (nécessaire pour déchiffrer)
    chiffré := gcm.Seal(nonce, nonce, texte, nil)
    return chiffré, nil
}

// DéchiffrerAESGCM déchiffre des données chiffrées avec ChiffrerAESGCM
func DéchiffrerAESGCM(données, cle []byte) ([]byte, error) {
    block, err := aes.NewCipher(cle)
    if err != nil {
        return nil, fmt.Errorf("création cipher : %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("création GCM : %w", err)
    }

    nonceSize := gcm.NonceSize()
    if len(données) < nonceSize {
        return nil, fmt.Errorf("données trop courtes")
    }

    nonce, chiffré := données[:nonceSize], données[nonceSize:]
    texte, err := gcm.Open(nil, nonce, chiffré, nil)
    if err != nil {
        return nil, fmt.Errorf("déchiffrement échoué (données corrompues ?) : %w", err)
    }
    return texte, nil
}

func main() {
    // Clé AES-256 — 32 bytes aléatoires
    cle := make([]byte, 32)
    rand.Read(cle)

    message := []byte("métriques confidentielles : CPU=87%")

    chiffré, err := ChiffrerAESGCM(message, cle)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Chiffré (%d bytes) : %x...\n", len(chiffré), chiffré[:8])

    déchiffré, err := DéchiffrerAESGCM(chiffré, cle)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Déchiffré : %s\n", déchiffré)
}
```

> ⚠️ **AES-GCM, pas AES-CBC** — Utilisez toujours AES-GCM (mode authentifié) plutôt qu'AES-CBC. AES-GCM détecte toute modification des données chiffrées (intégrité + confidentialité). AES-CBC ne garantit que la confidentialité et est vulnérable à des attaques comme POODLE.

> ⚠️ **Jamais de clés hardcodées** — Les clés de chiffrement viennent toujours de variables d'environnement, de secrets managers (Vault, AWS Secrets Manager), ou de fichiers protégés. Jamais dans le code source.

---

## 🛠️ Projet fil rouge — `gowatch` Pro : Scanner et export TLS

On intègre le scanner de ports et l'export sécurisé dans `gowatch`.

```go
package main

import (
    "context"
    "crypto/tls"
    "encoding/json"
    "flag"
    "fmt"
    "net"
    "net/http"
    "os"
    "runtime"
    "sort"
    "sync"
    "time"
)

// --- Types ---

type ResultatPort struct {
    Port   int  `json:"port"`
    Ouvert bool `json:"ouvert"`
}

type Metrique struct {
    Source string  `json:"source"`
    Valeur float64 `json:"valeur"`
    Unite  string  `json:"unite"`
}

type Snapshot struct {
    Timestamp time.Time  `json:"timestamp"`
    Metriques []Metrique `json:"metriques"`
    Ports     []ResultatPort `json:"ports,omitempty"`
}

// --- Scanner ---

func testerPort(ctx context.Context, hote string, port int) ResultatPort {
    var d net.Dialer
    conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", hote, port))
    if err != nil {
        return ResultatPort{Port: port, Ouvert: false}
    }
    conn.Close()
    return ResultatPort{Port: port, Ouvert: true}
}

func scanner(ctx context.Context, hote string, debut, fin int) []ResultatPort {
    resultats := make(chan ResultatPort, fin-debut+1)
    sem := make(chan struct{}, 200)
    var wg sync.WaitGroup

    for port := debut; port <= fin; port++ {
        wg.Add(1)
        p := port
        go func() {
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            portCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
            defer cancel()
            resultats <- testerPort(portCtx, hote, p)
        }()
    }

    go func() { wg.Wait(); close(resultats) }()

    var tous []ResultatPort
    for r := range resultats {
        if r.Ouvert {
            tous = append(tous, r)
        }
    }
    sort.Slice(tous, func(i, j int) bool { return tous[i].Port < tous[j].Port })
    return tous
}

// --- Export sécurisé ---

func exporterViaHTTPS(url string, snap Snapshot, skipVerify bool) error {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion:         tls.VersionTLS12,
            InsecureSkipVerify: skipVerify, // true uniquement pour les tests !
        },
    }
    client := &http.Client{Transport: transport, Timeout: 10 * time.Second}

    donnees, err := json.Marshal(snap)
    if err != nil {
        return fmt.Errorf("sérialisation : %w", err)
    }

    resp, err := client.Post(url, "application/json",
        bytes.NewReader(donnees))
    if err != nil {
        return fmt.Errorf("envoi : %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 300 {
        return fmt.Errorf("serveur a retourné %d", resp.StatusCode)
    }
    return nil
}

// --- Main ---

func main() {
    // Flags CLI
    scanHote   := flag.String("scan", "", "Hôte à scanner")
    scanPorts  := flag.String("ports", "1-1024", "Plage de ports (ex: 1-1024)")
    exportURL  := flag.String("export", "", "URL HTTPS d'export des métriques")
    skipVerify := flag.Bool("skip-verify", false, "Ignorer la vérification TLS (tests uniquement)")
    flag.Parse()

    ctx := context.Background()

    // Collecter les métriques
    snap := Snapshot{
        Timestamp: time.Now(),
        Metriques: []Metrique{
            {Source: "CPU",        Valeur: float64(runtime.NumCPU()),       Unite: "cœurs"},
            {Source: "Goroutines", Valeur: float64(runtime.NumGoroutine()), Unite: "actives"},
            {Source: "Go",         Valeur: 0, Unite: runtime.Version()},
        },
    }

    // Scanner si demandé
    if *scanHote != "" {
        var debut, fin int
        fmt.Sscanf(*scanPorts, "%d-%d", &debut, &fin)

        fmt.Printf("Scan de %s (ports %d-%d)...\n", *scanHote, debut, fin)
        scanDebut := time.Now()
        snap.Ports = scanner(ctx, *scanHote, debut, fin)
        fmt.Printf("%d ports ouverts trouvés en %v\n\n",
            len(snap.Ports), time.Since(scanDebut).Round(time.Millisecond))

        for _, p := range snap.Ports {
            fmt.Printf("  ✓ Port %d\n", p.Port)
        }
    }

    // Afficher les métriques
    fmt.Println("\n=== Métriques système ===")
    for _, m := range snap.Metriques {
        fmt.Printf("  %-12s : %.0f %s\n", m.Source, m.Valeur, m.Unite)
    }

    // Exporter si demandé
    if *exportURL != "" {
        fmt.Printf("\nExport vers %s...\n", *exportURL)
        if err := exporterViaHTTPS(*exportURL, snap, *skipVerify); err != nil {
            fmt.Fprintf(os.Stderr, "Erreur d'export : %v\n", err)
            os.Exit(1)
        }
        fmt.Println("Export réussi.")
    }
}
```

**Utilisation :**
```bash
# Scanner les ports locaux
go run main.go --scan localhost --ports 1-1024

# Collecter et exporter vers gohub
go run main.go --export https://localhost:8443/api/metrics/ingest --skip-verify

# Les deux ensemble
go run main.go --scan localhost --ports 80-9000 \
               --export https://gohub.example.com/api/metrics/ingest
```

---

## Ce qu'il faut retenir

1. **`net.Conn` est `io.Reader` + `io.Writer`** — une connexion TCP s'utilise exactement comme un fichier ou un buffer. Toutes les fonctions qui acceptent ces interfaces fonctionnent sur le réseau.

2. **Toujours des timeouts sur le réseau** — `DialTimeout`, `SetDeadline`. Une connexion sans timeout est une goroutine potentiellement bloquée pour toujours.

3. **Le sémaphore `chan struct{}`** — pour limiter la concurrence sans framework. Simple, efficace, idiomatique.

4. **AES-GCM pour le chiffrement symétrique** — authentifié, moderne, résistant aux attaques courantes. Pas AES-CBC.

5. **`crypto/tls` sans OpenSSL** — Go implémente TLS en Go pur. Moins de surface d'attaque, déploiement simplifié, et une configuration explicite qui vous oblige à faire des choix conscients.

---

## Pour aller plus loin

- 📄 [net — Documentation officielle](https://pkg.go.dev/net)
- 📄 [crypto/tls — Documentation officielle](https://pkg.go.dev/crypto/tls)
- 📄 [Cryptography in Go](https://pkg.go.dev/crypto) — Vue d'ensemble des primitives
- 🔧 [mTLS en Go](https://github.com/nicholasgasior/golang-mutual-tls) — Authentification mutuelle TLS

---

<div align="center">

[⬅️ Retour au Module 05](./README.md) · [👉 Chapitre 5.2 — Optimisation Bas Niveau](./02-optimisation-bas-niveau.md)

</div>
