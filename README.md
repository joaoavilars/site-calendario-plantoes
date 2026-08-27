# Calendário de Plantões

Sistema web para gerenciar a escala de plantões de suporte, substituindo a
planilha mensal do Google Sheets. Calendário mensal visual, cadastro de
integrantes, rodízio configurável (incluindo escalas alternadas como "sábado
sim / sábado não"), substituições e trocas com confirmação prévia, total de
horas mensais por integrante, histórico de alterações e alertas automáticos
via bot do Telegram.

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go (Gin) — API REST + serve o frontend embutido no binário |
| Frontend | React + Vite + TypeScript + Tailwind CSS (pt-BR) |
| Banco | SQLite (arquivo único em volume Docker, WAL) |
| Alertas | Bot do Telegram via HTTP (`sendMessage`) com scheduler cron |
| Deploy | Docker (multi-stage, imagem final ~35MB) |

## Início rápido

```bash
cp config/config.example.yaml config/config.yaml   # edite token/chat_id do Telegram
docker compose up -d --build
```

Acesse **http://localhost:8080**. Os dados ficam em `./data/plantoes.db`.

Primeiros passos na interface:
0. **Login** (topo da página) → usuário inicial `admin` / `admin123`; troque
   a senha na engrenagem ⚙️. Sem login, o calendário fica somente leitura.
1. **Integrantes** → cadastre cada pessoa (nome, cor, @menção do Telegram).
2. **Rodízio** → defina os dias da semana e horários de plantão de cada um.
3. **Calendário** → visualize a escala; clique num dia para registrar
   substituição, troca, plantão extra ou ausência (sempre com prévia e
   confirmação antes de valer).

## Desenvolvimento local

```bash
# Terminal 1 — backend (porta 8080)
cd backend && go run ./cmd/server

# Terminal 2 — frontend com hot-reload (porta 5173, proxy /api → 8080)
cd frontend && npm install && npm run dev
```

Testes do motor de escala (reproduzem a planilha original — 62h/62h):

```bash
cd backend && go test ./...
```

## Documentação

- [docs/arquitetura.md](docs/arquitetura.md) — componentes, modelo de dados e o motor de materialização
- [docs/api.md](docs/api.md) — todas as rotas da API com exemplos `curl`
- [docs/manual-de-uso.md](docs/manual-de-uso.md) — passo a passo do gestor + configuração do bot do Telegram
- [docs/deploy.md](docs/deploy.md) — deploy, backup e atualização

## Estrutura

```
backend/     API Go: internal/schedule (motor), store, api, notify (Telegram)
frontend/    SPA React (buildada e embutida no binário Go)
config/      config.yaml (token do bot, horários dos alertas, fuso)
data/        banco SQLite (volume Docker)
docs/        documentação
```
