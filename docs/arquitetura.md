# Arquitetura

## Visão geral

Um único container roda um binário Go que serve a API REST (`/api/...`) e o
frontend React embutido via `go:embed`. O SQLite fica num volume (`/data`), e
um scheduler interno (robfig/cron) envia os alertas do Telegram.

```
┌────────────────────────── container ──────────────────────────┐
│  binário Go                                                   │
│  ├── Gin: /api/* (REST)  +  /* (SPA React embutida)           │
│  ├── internal/schedule: motor de materialização do calendário │
│  ├── internal/store: SQLite (/data/plantoes.db, WAL)          │
│  └── internal/notify: cron + Telegram sendMessage             │
└───────────────────────────────────────────────────────────────┘
```

## Conceito central: o calendário nunca é armazenado

A escala de cada mês é **computada sob demanda** pelo motor
(`backend/internal/schedule/engine.go`) a partir de duas fontes:

1. **Regras de rodízio** (`rotation_rules`) — recorrência semanal por
   integrante: dia da semana, horário de início/fim, e `interval_weeks`
   (1 = toda semana; 2 = alternado, ex. sábado sim/sábado não, ancorado na
   data `effective_from`). Cada regra tem vigência (`effective_from` /
   `effective_to`): editar o rodízio **encerra** as regras atuais e cria
   novas, então meses passados continuam materializando exatamente como
   eram — é isso que dá o histórico confiável.

2. **Exceções** (`exceptions`) — agendamentos individuais por data:
   - `substituicao`: os trechos do plantão do integrante original dentro da
     faixa informada são transferidos ao substituto (a estrutura dos slots é
     preservada — um plantão de sábado com pausa de almoço não "ganha" a
     hora do almoço).
   - `troca`: substituição na data original **mais** a inversa na
     `swap_date` (o original assume os plantões do substituto naquele dia).
   - `extra`: adiciona um plantão individual avulso.
   - `ausencia`: remove o plantão; se ninguém cobre a faixa, o dia é
     marcado como **descoberto** (⚠️ no calendário).
   - `feriado` (dia inteiro): cancela toda a escala automática do dia
     (🎉 no calendário, com o motivo); aplicado **antes** das exceções
     individuais, então um plantão `extra` confirmado depois ainda vale.
   - `edicao_dia` (dia inteiro): substitui a escala do dia pelos slots
     gravados em `slots_json` (origem `edicao`) — usado para dias
     diferenciados, ex. véspera de Natal com escala de sábado.
   - `ferias` (período `date`→`end_date`): mascara os plantões do integrante
     no intervalo — com substituto, os slots são transferidos a ele (origem
     `ferias`); sem substituto, os dias afetados ficam descobertos. As
     regras do rodízio continuam correndo por baixo, então no retorno a
     escala (inclusive alternâncias) volta exatamente de onde parou.

   Exceções nascem com status `pendente` e **não afetam a escala nem as
   horas** até serem confirmadas. A API devolve uma prévia (antes/depois +
   delta de horas) na criação, e o frontend só aplica após o clique em
   "Confirmar alteração". Pendentes aparecem tracejadas no calendário.

As **horas mensais** de cada integrante são a soma dos slots materializados
no mês — o mesmo número que o rodapé da planilha antiga calculava (o teste
`engine_test.go` reproduz setembro/2026 da planilha: 62h para cada um).

## Modelo de dados (SQLite)

| Tabela | Papel |
|---|---|
| `members` | Integrantes (nome, cor, @menção Telegram, ativo) |
| `rotation_rules` | Regras semanais com vigência e `interval_weeks` |
| `exceptions` | Substituições/trocas/extras/ausências com fluxo `pendente → confirmado/rejeitado/cancelado` |
| `audit_log` | Toda mutação (quem/quê/quando + snapshot JSON) — alimenta a aba Histórico |
| `alerts_sent` | Chave única por alerta enviado — evita duplicatas em restart |

Migrações: `pressly/goose`, arquivos SQL embutidos (`backend/migrations/`),
aplicadas automaticamente no boot. Driver: `modernc.org/sqlite` (100% Go,
sem CGO), com `journal_mode=WAL` e `busy_timeout=5000`.

## Alertas do Telegram (`internal/notify`)

| Alerta | Gatilho | Dedup (`alerts_sent`) |
|---|---|---|
| Resumo diário | cron no horário do `config.yaml` | `resumo:AAAA-MM-DD` |
| Lembrete de plantão | job por minuto: slots começando em `agora + antecedência` | `lembrete:data:membro:hora` |
| Troca confirmada | direto no handler de confirmação | — (evento único) |

Envio: `POST https://api.telegram.org/bot<token>/sendMessage` com timeout e
uma retentativa. Sem token/chat_id configurados, as mensagens vão para o log
(o sistema funciona normalmente sem Telegram).

O cron roda no fuso do `config.yaml` (`America/Sao_Paulo`); a imagem final
inclui `tzdata` para isso.

## Decisões e limitações

- **Autenticação**: login administrativo único (tabela `admins`, bcrypt)
  com sessões por token opaco (tabela `sessions`, validade de 7 dias,
  header `Authorization: Bearer`). Público: só a visualização do calendário
  (`GET /api/calendario`), o health check e o login — todo o resto exige
  sessão. Para expor à internet, adicione HTTPS via reverse proxy.
- **Plantões não cruzam a meia-noite** (início < fim no mesmo dia). Um
  plantão noturno 22h–06h precisaria ser modelado como dois blocos.
- **SQLite com 1 conexão** — suficiente para uma equipe; o gargalo real é
  inexistente nesse volume.
- Frontend sem router/Redux — abas em estado local, fetch simples.
