# Calendário de Plantões

Sistema web para gerenciar a escala de plantões de suporte, substituindo a
planilha mensal do Google Sheets. Calendário mensal visual, cadastro de
integrantes, rodízio configurável (incluindo escalas alternadas como "sábado
sim / sábado não"), substituições, trocas, plantões avulsos, feriados, dias
com escala manualmente ajustada e férias — tudo com prévia e confirmação
antes de valer. Total de horas mensais por integrante em destaque, histórico
de alterações, alertas automáticos via bot do Telegram e login
administrativo protegendo a edição.

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go (Gin) — API REST + serve o frontend embutido no binário |
| Frontend | React + Vite + TypeScript + Tailwind CSS (pt-BR, tema escuro por padrão) |
| Banco | SQLite (arquivo único em volume Docker, WAL) |
| Autenticação | Login administrativo único, sessão por token, senha com bcrypt |
| Alertas | Bot do Telegram via HTTP (`sendMessage`) com scheduler cron |
| Deploy | Docker (multi-stage, imagem final ~35MB) |

## Início rápido

```bash
cp config/config.example.yaml config/config.yaml   # edite token/chat_id do Telegram (opcional)
docker compose up -d --build
```

Acesse **http://localhost:8080**. Os dados ficam em `./data/plantoes.db`.

### 🔑 Acesso administrativo

| Usuário | Senha |
|---|---|
| `admin` | `admin123` |

**Troque essa senha assim que possível**, no ícone ⚙️ no canto superior
direito (Configurações do administrador → só permite alterar a senha, exige
a atual e mínimo de 6 caracteres). Sem login, o site mostra apenas o
calendário em modo leitura — nenhuma edição fica disponível nem visível.

## Como usar

O acesso funciona em duas camadas:

- **Sem login**: qualquer pessoa com o link vê o calendário do mês, com
  quem está escalado, o horário de cada um e avisos de feriado, férias ou
  horário descoberto. Não é possível clicar nos dias nem ver as demais
  telas.
- **Logado como administrador** (campos de usuário/senha no topo da
  página): liberam as abas **Integrantes**, **Rodízio** e **Histórico**, e
  o clique num dia do calendário abre o modal de edição.

Passo a passo típico do gestor:

1. **👥 Integrantes** — cadastre cada pessoa da equipe (nome, cor no
   calendário, @menção do Telegram opcional para os alertas). Desativar um
   integrante encerra o rodízio dele a partir de hoje, sem apagar o
   histórico.
2. **🔁 Rodízio** — defina os plantões de cada integrante: escolha uma
   faixa de dias (Seg a Sex, Seg a Sáb, todos os dias, ou um dia
   específico), o horário e se é toda semana ou alternado (ex.: sábado sim
   / sábado não, informando a partir de qual data começa). Nesta mesma
   aba fica a seção **🏖️ Férias e afastamentos**: informe o período e,
   opcionalmente, quem cobre — ao voltar das férias, a escala retoma
   sozinha de onde parou, sem precisar reconfigurar nada.
3. **📅 Calendário** — visualize a escala do mês, com o total de horas de
   cada integrante em destaque no topo. Clique num dia para:
   - **Substituição** ou **Troca** de plantão entre dois integrantes;
   - **Plantão individual** avulso, ou **Ausência** (fica marcado como
     horário descoberto);
   - **Sem plantão (feriado)** — cancela a escala automática só daquele
     dia (Natal, Ano Novo etc.);
   - **Editar o dia** — redefine manualmente toda a escala de um dia
     específico (ex.: uma quarta-feira que precisa funcionar como sábado).

   Toda alteração mostra uma prévia (antes → depois, com o impacto nas
   horas do mês) antes de pedir confirmação — nada é aplicado sem esse
   passo.
4. **📜 Histórico** — auditoria de tudo que foi criado, alterado ou
   confirmado, com filtros por tipo e período.
5. **⚙️ Configurações** — troca de senha do administrador.

### Alertas no Telegram (opcional)

Preenchendo `bot_token` e `chat_id` em `config/config.yaml`, o sistema
envia automaticamente: resumo diário da escala pela manhã, lembrete alguns
minutos antes de cada plantão, e aviso quando uma troca/substituição/férias
é confirmada. Veja o passo a passo (criar o bot no BotFather, obter o
chat_id) em [docs/manual-de-uso.md](docs/manual-de-uso.md).

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
- [docs/api.md](docs/api.md) — todas as rotas da API com exemplos `curl` (incluindo autenticação)
- [docs/manual-de-uso.md](docs/manual-de-uso.md) — passo a passo do gestor + configuração do bot do Telegram
- [docs/deploy.md](docs/deploy.md) — deploy, backup, segurança e atualização

## Estrutura

```
backend/     API Go: internal/schedule (motor), store, api, notify (Telegram)
frontend/    SPA React (buildada e embutida no binário Go)
config/      config.yaml (token do bot, horários dos alertas, fuso)
data/        banco SQLite (volume Docker)
docs/        documentação
```
