# Deploy, backup e atualização

## Subir

```bash
cp config/config.example.yaml config/config.yaml   # edite o Telegram
docker compose up -d --build
```

- Porta: **8080** (ajuste o mapeamento em `docker-compose.yml` se precisar).
- Fuso horário dos alertas: `timezone` no `config.yaml` (padrão
  `America/Sao_Paulo`) — vale para o horário do resumo diário e lembretes.
- O container reinicia sozinho (`restart: unless-stopped`) e tem healthcheck
  em `/api/healthz`.

## Onde ficam os dados

| O quê | Onde |
|---|---|
| Banco SQLite | `./data/plantoes.db` (+ `-wal`/`-shm`) |
| Configuração | `./config/config.yaml` (montado somente leitura) |

## Backup

O banco é um arquivo único. Backup a frio (recomendado, segundos de pausa):

```bash
docker compose stop && cp data/plantoes.db /destino/backup-$(date +%F).db && docker compose start
```

A quente também funciona na prática para esse volume de uso (copie os três
arquivos `plantoes.db*` juntos). Restaurar = devolver o arquivo para `data/`
e reiniciar.

## Atualizar a aplicação

```bash
docker compose up -d --build
```

As migrações de banco rodam automaticamente no boot — não há passo manual.

## Sem duplicar alertas

Os alertas enviados ficam registrados na tabela `alerts_sent`; reiniciar o
container no meio do dia **não** reenvia o resumo diário nem lembretes já
enviados.

## Logs

```bash
docker compose logs -f
```

Sem `bot_token`/`chat_id` configurados, os alertas aparecem apenas no log
(prefixo `[telegram] desabilitado`) — útil para validar horários antes de
ligar o bot de verdade.

## Segurança

- **Login administrativo**: a edição (integrantes, rodízio, férias,
  agendamentos, histórico) exige login. Sem login, só a visualização do
  calendário fica acessível. Usuário inicial **admin / admin123** — troque a
  senha na engrenagem ⚙️ logo no primeiro acesso. Senhas com bcrypt;
  sessões de 7 dias na tabela `sessions`.
- O sistema foi pensado para rede interna. Para expor à internet, além do
  login, coloque HTTPS na frente (reverse proxy — Nginx/Caddy) — o token de
  sessão trafega no header e não deve viajar em HTTP aberto.
