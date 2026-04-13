VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE     := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS  := -ldflags="-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(DATE)"

.PHONY: build test run watch scan clean

## build : Compiler le binaire pour l'OS courant
build:
	go build $(LDFLAGS) -o gowatch .

## test : Lancer les tests avec race detector
test:
	go test -race -cover ./...

## run : Lancer un snapshot unique (mode texte)
run:
	go run . 

## watch : Lancer en mode surveillance toutes les 2 secondes
watch:
	go run . --watch --interval 2s

## scan : Scanner les ports 1-1024 de localhost
scan:
	go run . --scan localhost --ports 1-1024

## json : Snapshot en JSON
json:
	go run . --format json

## linux : Compiler pour Linux AMD64 (déploiement serveur)
linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/gowatch-linux-amd64 .

## all-platforms : Compiler pour toutes les plateformes
all-platforms:
	@mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build $(LDFLAGS) -o dist/gowatch-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build $(LDFLAGS) -o dist/gowatch-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build $(LDFLAGS) -o dist/gowatch-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build $(LDFLAGS) -o dist/gowatch-darwin-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/gowatch-windows.exe .
	@echo "Binaires dans dist/ :"
	@ls -lh dist/

## clean : Supprimer les artefacts de build
clean:
	rm -f gowatch
	rm -rf dist/

## help : Afficher l'aide
help:
	@grep -E '^## ' Makefile | sed 's/## //'
