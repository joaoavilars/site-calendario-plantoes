# API REST

Base: `http://localhost:8080/api`. Todas as respostas são JSON. Erros vêm
como `{"erro": "mensagem"}` com status 4xx/5xx.

## Autenticação

Somente `GET /calendario/:ano/:mes`, `GET /healthz` e `POST /login` são
públicos. Todas as demais rotas exigem o token de sessão no header
`Authorization: Bearer <token>` (sem ele: HTTP 401).

```bash
# Login (usuário inicial: admin / admin123 — troque após o primeiro acesso)
TOKEN=$(curl -s -X POST localhost:8080/api/login -H 'Content-Type: application/json' \
  -d '{"usuario":"admin","senha":"admin123"}' | jq -r .token)

# Usar o token
curl -H "Authorization: Bearer $TOKEN" localhost:8080/api/membros

# Sessão atual / logout / troca de senha (mínimo 6 caracteres)
curl -H "Authorization: Bearer $TOKEN" localhost:8080/api/me
curl -X POST -H "Authorization: Bearer $TOKEN" localhost:8080/api/logout
curl -X POST -H "Authorization: Bearer $TOKEN" localhost:8080/api/senha \
  -H 'Content-Type: application/json' \
  -d '{"senha_atual":"admin123","senha_nova":"nova-senha"}'
```

Sessões valem por 7 dias; trocar a senha desconecta as demais sessões. As
senhas são armazenadas com bcrypt.

## Integrantes

```bash
# Listar ativos (todos=1 inclui inativos)
curl localhost:8080/api/membros
curl "localhost:8080/api/membros?todos=1"

# Criar
curl -X POST localhost:8080/api/membros -H 'Content-Type: application/json' \
  -d '{"nome":"João","cor":"#22c55e","mencao_telegram":"@joao"}'

# Atualizar (campos omitidos são preservados; ativo reativa/desativa)
curl -X PUT localhost:8080/api/membros/1 -H 'Content-Type: application/json' \
  -d '{"nome":"João Silva","ativo":true}'

# Desativar (soft delete: encerra também o rodízio a partir de hoje)
curl -X DELETE localhost:8080/api/membros/1
```

## Rodízio

```bash
# Regras vigentes do integrante (historico=1 lista todas)
curl localhost:8080/api/membros/1/regras

# Substituir o rodízio do integrante a partir de vigente_de (padrão: hoje).
# As regras antigas são encerradas na véspera — o histórico é preservado.
# dia_semana: 0=domingo ... 6=sábado
# intervalo_semanas: 1=toda semana; 2=alternado (vigente_de da regra ancora a alternância)
curl -X PUT localhost:8080/api/membros/1/regras -H 'Content-Type: application/json' -d '{
  "vigente_de": "2026-01-01",
  "regras": [
    {"dia_semana": 1, "inicio": "06:00", "fim": "08:00"},
    {"dia_semana": 6, "intervalo_semanas": 2, "vigente_de": "2026-09-12",
     "inicio": "07:00", "fim": "12:00"}
  ]
}'

# Encerrar uma regra específica (a partir de hoje)
curl -X DELETE localhost:8080/api/regras/3
```

## Calendário

```bash
curl localhost:8080/api/calendario/2026/9
```

Resposta (resumida):

```json
{
  "ano": 2026, "mes": 9,
  "dias": [
    {
      "data": "2026-09-14",
      "slots": [
        {"membro_id": 2, "nome": "Gustavo", "cor": "#eab308",
         "inicio": "06:00", "fim": "08:00", "origem": "substituicao"}
      ],
      "pendentes": [],
      "descoberto": false
    }
  ],
  "horas": [
    {"membro_id": 1, "nome": "João", "cor": "#22c55e",
     "total_minutos": 3600, "total_formatado": "60h00"}
  ]
}
```

`origem`: `rodizio` | `substituicao` | `troca` | `extra`.

## Agendamentos individuais (exceções)

```bash
# Criar (nasce "pendente"; a resposta traz a prévia antes/depois + delta de horas)
curl -X POST localhost:8080/api/excecoes -H 'Content-Type: application/json' -d '{
  "data": "2026-09-14",
  "tipo": "substituicao",
  "membro_original_id": 1,
  "membro_substituto_id": 2,
  "inicio": "06:00", "fim": "08:00",
  "observacao": "consulta médica"
}'
```

Tipos e campos obrigatórios:

| tipo | campos |
|---|---|
| `substituicao` | `membro_original_id`, `membro_substituto_id`, `inicio`, `fim` |
| `troca` | os dois acima + `data_troca` (dia em que o original assume os plantões do substituto) |
| `extra` | `membro_substituto_id` (quem fará o plantão individual/avulso) |
| `ausencia` | `membro_original_id` (o dia pode ficar `descoberto`) |
| `feriado` | nenhum — cancela toda a escala automática do dia; use `observacao` para o motivo (ex.: "Natal") |
| `edicao_dia` | `slots`: nova escala completa do dia, ex. `[{"membro_id":1,"inicio":"08:00","fim":"12:00"}]` |
| `ferias` | `membro_original_id`, `data_fim` (período `data`→`data_fim`, máx. 90 dias); `membro_substituto_id` opcional assume os plantões no período |

Exemplos dos tipos de dia inteiro:

```bash
# Feriado — dia sem plantão
curl -X POST localhost:8080/api/excecoes -H 'Content-Type: application/json' \
  -d '{"data":"2026-12-25","tipo":"feriado","observacao":"Natal"}'

# Dia diferenciado — 24/12 (quinta) com escala de sábado
curl -X POST localhost:8080/api/excecoes -H 'Content-Type: application/json' -d '{
  "data": "2026-12-24", "tipo": "edicao_dia",
  "observacao": "véspera de Natal",
  "slots": [
    {"membro_id": 1, "inicio": "08:00", "fim": "12:00"},
    {"membro_id": 2, "inicio": "13:00", "fim": "17:00"}
  ]
}'
```

```bash
# Férias de João (14–20/09) com Gustavo cobrindo os plantões
curl -X POST localhost:8080/api/excecoes -H 'Content-Type: application/json' -d '{
  "data": "2026-09-14", "data_fim": "2026-09-20", "tipo": "ferias",
  "membro_original_id": 1, "membro_substituto_id": 2
}'

# Listar férias de um integrante (filtros tipo/membro valem para qualquer exceção)
curl "localhost:8080/api/excecoes?tipo=ferias&membro=1"
```

No calendário, um dia com `feriado` confirmado vem com `sem_plantao: true` e
`motivo`; slots de `edicao_dia` têm `origem: "edicao"`; dias dentro de férias
listam quem está fora em `ferias: [{membro_id, nome}]` e slots cobertos pelo
substituto têm `origem: "ferias"`.

```bash
# Listar (filtros opcionais: ano, mes, status)
curl "localhost:8080/api/excecoes?ano=2026&mes=9&status=pendente"

# Confirmar (aplica na escala e dispara o aviso no Telegram)
curl -X POST localhost:8080/api/excecoes/1/confirmar

# Rejeitar (pendente) / Cancelar (pendente ou confirmada)
curl -X POST localhost:8080/api/excecoes/1/rejeitar
curl -X DELETE localhost:8080/api/excecoes/1
```

## Histórico e utilitários

```bash
# Auditoria (filtros: entidade=member|rule|exception, de/ate=YYYY-MM-DD, limite)
curl "localhost:8080/api/historico?entidade=exception&de=2026-09-01"

# Testa o envio ao grupo do Telegram
curl -X POST localhost:8080/api/alertas/teste

# Health check
curl localhost:8080/api/healthz
```
