# Manual de uso

## 0. Login do administrador

Sem login, o site mostra **apenas o calendário, somente leitura** — sem
modal de edição e sem as abas de gestão. Os campos de usuário/senha ficam no
topo da página, com o botão **Entrar**.

- Primeiro acesso: **admin / admin123**. Troque a senha imediatamente na
  engrenagem **⚙️** (canto superior direito), que abre as configurações do
  administrador — lá é possível alterar a senha (mínimo 6 caracteres).
- Ao logar, as abas **Integrantes**, **Rodízio** e **Histórico** aparecem e
  o calendário passa a abrir o modal de edição ao clicar num dia.
- **Sair** encerra a sessão; sessões expiram sozinhas em 7 dias; trocar a
  senha desconecta as demais sessões abertas.

## 1. Cadastrar integrantes

Aba **👥 Integrantes** → preencha nome, escolha uma cor (é a cor dos blocos
no calendário) e, opcionalmente, a menção do Telegram (`@usuario` — usada nos
alertas do grupo). Clique em **Cadastrar**.

Para remover alguém do quadro, use **Desativar**: o rodízio da pessoa é
encerrado a partir de hoje, mas os meses passados continuam registrados
(histórico preservado). **Reativar** traz a pessoa de volta (configure o
rodízio novamente).

## 2. Configurar o rodízio

Aba **🔁 Rodízio** → selecione o integrante → **+ Adicionar plantão** para
cada bloco:

- Escolha uma **faixa de dias** (Seg a Sex, Seg a Sáb, Sáb e Dom, Todos os
  dias) ou um dia específico — a faixa cria o plantão em todos os dias de
  uma vez com o mesmo horário.
- Horário de início e fim ao lado.
- **"toda semana"** para plantões fixos.
- **"a cada 2 semanas"** para escalas alternadas (ex.: sábado sim / sábado
  não). Informe em **"começando em"** a data do primeiro sábado daquela
  pessoa — é isso que define quem pega cada sábado.
- Plantões com pausa (ex.: sábado 07–17 com almoço) entram como **duas
  linhas**: 07:00–12:00 e 13:00–17:00.

**Salvar rodízio** substitui as regras vigentes a partir de hoje; o passado
não muda. O calendário e as horas se recalculam na hora.

### Exemplo — reproduzir a planilha antiga

- João: faixa "Seg a Sex" 06:00–08:00 (toda semana) + Sábado 07:00–12:00 e
  13:00–17:00 (a cada 2 semanas, começando no sábado dele).
- Gustavo: seg–sex 18:00–20:00 (toda semana) + os mesmos blocos de sábado
  (a cada 2 semanas, começando no sábado alternado).

Resultado: 62h/62h no mês, igual ao rodapé da planilha.

## 2.1 Férias e afastamentos

Na mesma aba **🔁 Rodízio**, abaixo do rodízio do integrante, a seção
**🏖️ Férias e afastamentos**:

1. Informe o período (De/Até) e, opcionalmente, um **substituto** — que
   assume automaticamente todos os plantões do integrante no período, com as
   horas contando para ele. Sem substituto, os dias ficam sem o plantão e
   marcados como descobertos ⚠️ (cubra pontualmente com "Plantão individual"
   no calendário, se precisar).
2. **Agendar férias** mostra a prévia (dias afetados + horas antes → depois);
   nada vale até **Confirmar férias**.
3. Cada dia do período aparece no calendário com a badge **🏖️ Férias: Nome**.
4. O rodízio **continua correndo por baixo**: no retorno, a escala — inclusive
   alternâncias de sábado — volta exatamente de onde parou, sem mexer nas
   regras. Cancelar um período (botão Cancelar na lista) devolve os plantões
   imediatamente.

## 3. Calendário e horas

Aba **📅 Calendário**: cada dia mostra quem cobre e o horário, na cor do
integrante. No topo, o **total de horas do mês por integrante** — a visão
rápida do gestor sobre a distribuição de carga. O total considera **apenas
os dias do mês exibido** (nada é acumulado de meses anteriores); meses com
mais sábados, por exemplo, naturalmente somam mais horas.

O tema escuro é o padrão; o botão no canto superior direito alterna para o
claro (a preferência fica salva no navegador).

Legenda: **✱** alteração confirmada (substituição/troca/extra) · **⏳**
pendente de confirmação · **⚠️** horário descoberto.

## 4. Substituições, trocas, extras e ausências

Clique no dia desejado no calendário:

1. Escolha o tipo, os integrantes e a faixa de horário:
   - **Substituição** — outro integrante assume o plantão no dia.
   - **Troca** — dois integrantes trocam plantões entre duas datas.
   - **Plantão individual** — inclui um plantão avulso para alguém no dia.
   - **Ausência** — remove o plantão (o dia pode ficar descoberto ⚠️).
   - **Sem plantão (feriado)** — cancela toda a escala automática só daquele
     dia (Natal, Ano Novo etc.) e recalcula as horas do mês; informe o
     motivo, que aparece no calendário (🎉). Plantões individuais
     adicionados depois continuam valendo.
   - **Editar o dia** — redefine manualmente a escala inteira do dia: abre
     com os plantões atuais pré-carregados para ajustar, remover ou
     adicionar. Útil para dias diferenciados, ex.: 24/12 numa quarta-feira
     funcionando com escala de sábado.
2. Clique em **Ver prévia da alteração** — o sistema mostra o dia **antes e
   depois** e o efeito nas horas do mês. *Nada foi alterado ainda.*
3. **Confirmar alteração** aplica na escala (e avisa o grupo do Telegram, se
   configurado). **Descartar** joga fora.

Numa **troca**, a data selecionada é o plantão que o colega assume; em
"data em que quem entra será retribuído" informe o dia em que o titular
original assume os plantões do colega de volta.

Agendamentos pendentes de outro dia aparecem no topo do mesmo diálogo, com
botões Confirmar/Rejeitar.

## 5. Histórico

Aba **📜 Histórico**: toda alteração (cadastros, rodízio, agendamentos e suas
confirmações) fica registrada com data e detalhes (JSON), com filtros por
tipo e período.

## 6. Alertas do Telegram

### Criar o bot e obter o chat_id (uma vez só)

1. No Telegram, fale com **@BotFather** → `/newbot` → guarde o **token**.
2. Adicione o bot ao grupo do plantão.
3. Envie qualquer mensagem no grupo e abra
   `https://api.telegram.org/bot<TOKEN>/getUpdates` no navegador — o
   `"chat":{"id":-100...}` é o **chat_id** do grupo (número negativo).
4. Preencha `config/config.yaml` (`telegram.bot_token` e `telegram.chat_id`)
   e reinicie: `docker compose restart`.

Dica: para não deixar o token em arquivo, exporte `TELEGRAM_BOT_TOKEN` no
ambiente do `docker compose` (ele sobrepõe o do config).

### O que é enviado

| Alerta | Quando | Configuração |
|---|---|---|
| 📅 Resumo do dia | toda manhã | `alertas.resumo_diario.horario` (padrão 07:30) |
| ⏰ Lembrete | antes de cada plantão | `alertas.lembrete_plantao.antecedencia_minutos` (padrão 15) |
| 🔁 Confirmações | ao confirmar substituição/troca/extra/ausência | `alertas.aviso_troca_confirmada` |

Teste pelo botão **"Enviar mensagem de teste ao grupo"** na aba Integrantes
(ou `curl -X POST localhost:8080/api/alertas/teste`).
