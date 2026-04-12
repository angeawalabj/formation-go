# ─── Stage 1 : Compilation ───────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copier les dépendances en premier — optimise le cache Docker
COPY go.mod go.sum* ./
RUN go mod download

COPY . .

# Arguments de build injectés par docker build --build-arg
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildDate=${BUILD_DATE}" \
    -o /gowatch \
    .

# ─── Stage 2 : Image finale minimale ─────────────────────────────────────────
FROM scratch

# Certificats SSL pour les requêtes HTTPS sortantes
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Uniquement le binaire
COPY --from=builder /gowatch /gowatch

ENTRYPOINT ["/gowatch"]
