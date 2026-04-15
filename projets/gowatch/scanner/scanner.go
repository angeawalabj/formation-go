// Package scanner fournit un scanner de ports TCP concurrent.
// Il utilise un sémaphore pour limiter le nombre de connexions simultanées
// et le Context pour le support d'annulation et de timeout.
package scanner

import (
	"context"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// Résultat représente le résultat du test d'un port.
type Résultat struct {
	Port   int  `json:"port"`
	Ouvert bool `json:"ouvert"`
}

// Scan scanne la plage [portDebut, portFin] sur l'hôte donné.
// concurrence limite le nombre de connexions TCP simultanées.
// Retourne uniquement les ports ouverts, triés par numéro.
func Scan(ctx context.Context, hote string, portDebut, portFin, concurrence int) []Résultat {
	résultats := make(chan Résultat, portFin-portDebut+1)
	sémaphore  := make(chan struct{}, concurrence)

	var wg sync.WaitGroup

	for port := portDebut; port <= portFin; port++ {
		// Annulation anticipée si le contexte est expiré
		select {
		case <-ctx.Done():
			break
		default:
		}

		wg.Add(1)
		p := port
		go func() {
			defer wg.Done()

			sémaphore <- struct{}{}
			defer func() { <-sémaphore }()

			portCtx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
			defer cancel()

			résultats <- testerPort(portCtx, hote, p)
		}()
	}

	go func() {
		wg.Wait()
		close(résultats)
	}()

	var ouverts []Résultat
	for r := range résultats {
		if r.Ouvert {
			ouverts = append(ouverts, r)
		}
	}

	sort.Slice(ouverts, func(i, j int) bool {
		return ouverts[i].Port < ouverts[j].Port
	})

	return ouverts
}

// testerPort tente une connexion TCP sur le port donné.
// Retourne un Résultat avec Ouvert=true si la connexion réussit.
func testerPort(ctx context.Context, hote string, port int) Résultat {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", hote, port))
	if err != nil {
		return Résultat{Port: port, Ouvert: false}
	}
	conn.Close()
	return Résultat{Port: port, Ouvert: true}
}

// NomService retourne le nom conventionnel d'un port connu.
func NomService(port int) string {
	services := map[int]string{
		21: "FTP", 22: "SSH", 23: "Telnet", 25: "SMTP",
		53: "DNS", 80: "HTTP", 110: "POP3", 143: "IMAP",
		443: "HTTPS", 465: "SMTPS", 587: "SMTP/TLS",
		993: "IMAPS", 995: "POP3S", 3306: "MySQL",
		5432: "PostgreSQL", 6379: "Redis", 8080: "HTTP-Alt",
		8443: "HTTPS-Alt", 9000: "HTTP-Alt-2", 27017: "MongoDB",
	}
	if nom, ok := services[port]; ok {
		return nom
	}
	return "inconnu"
}
