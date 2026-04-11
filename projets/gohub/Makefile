VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE     := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  := -ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)"

.PHONY: build test bench run docker clean

## build : Compiler le binaire
build:
	go build $(LDFLAGS) -o gohub .

## test : Lancer les tests avec race detector et couverture
test:
	go test -race -cover ./...

## bench : Lancer les benchmarks
bench:
	go test -bench=. -benchmem ./...

## run : Démarrer le serveur (port 8080)
run:
	go run . --port 8080

## run-json : Démarrer avec logs JSON (mode production)
run-json:
	go run . --port 8080 --log-format json

## cover : Rapport de couverture HTML
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## docker : Construire l'image Docker
docker:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(DATE) \
		-t gohub:$(VERSION) \
		-t gohub:latest \
		.
	@echo "Image construite :"
	@docker images gohub:latest

## docker-run : Lancer le container
docker-run:
	docker run -d --name gohub -p 8080:8080 -e LOG_FORMAT=json gohub:latest
	@echo "gohub disponible sur http://localhost:8080"

## docker-stop : Arrêter et supprimer le container
docker-stop:
	docker stop gohub && docker rm gohub

## linux : Compiler pour Linux AMD64
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/gohub-linux-amd64 .

## clean : Supprimer les artefacts
clean:
	rm -f gohub coverage.out
	rm -rf dist/
	docker rmi gohub:latest 2>/dev/null || true

## curl-health : Tester l'endpoint /health
curl-health:
	curl -s http://localhost:8080/health | python3 -m json.tool

## curl-metrics : Tester l'endpoint /api/metrics
curl-metrics:
	curl -s http://localhost:8080/api/metrics | python3 -m json.tool

## curl-history : Tester l'historique (10 derniers)
curl-history:
	curl -s "http://localhost:8080/api/metrics/history?limit=10" | python3 -m json.tool

## help : Afficher l'aide
help:
	@grep -E '^## ' Makefile | sed 's/## //'
