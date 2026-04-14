// Package exporter envoie les snapshots vers une instance gohub via HTTPS.
package exporter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/votre-pseudo/gowatch/collector"
)

// ExporterHTTPS envoie un snapshot vers l'URL donnée via HTTP POST.
// Si skipVerify est true, la vérification du certificat TLS est désactivée.
// À n'utiliser qu'en environnement de test — jamais en production.
func ExporterHTTPS(url string, snap collector.Snapshot, skipVerify bool) error {
	donnees, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("ExporterHTTPS sérialisation : %w", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: skipVerify, //nolint:gosec // intentionnel pour les tests
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Post(url, "application/json", bytes.NewReader(donnees))
	if err != nil {
		return fmt.Errorf("ExporterHTTPS envoi : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("ExporterHTTPS : serveur a retourné HTTP %d", resp.StatusCode)
	}

	return nil
}
