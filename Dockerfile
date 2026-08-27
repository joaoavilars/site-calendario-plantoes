# Estágio 1: build do frontend React
FROM node:22-alpine AS frontend
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Estágio 2: build do backend Go (com frontend embutido)
FROM golang:1.25-alpine AS backend
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /app/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /plantoes ./cmd/server

# Estágio 3: imagem final mínima (ca-certificates para HTTPS do Telegram,
# tzdata para o fuso horário do scheduler de alertas)
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=backend /plantoes /usr/local/bin/plantoes
ENV DB_PATH=/data/plantoes.db \
    CONFIG_PATH=/config/config.yaml
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["plantoes"]
