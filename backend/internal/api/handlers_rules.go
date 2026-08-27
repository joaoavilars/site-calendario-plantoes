package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/store"
)

func (s *Server) listRules(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var (
		rules []store.Rule
		err   error
	)
	if c.Query("historico") == "1" {
		rules, err = s.St.ListRulesByMember(id, false, "")
	} else {
		// Inclui também regras com vigência futura (ex.: sábado alternado
		// ancorado numa data adiante) — senão elas aparecem no calendário
		// mas somem da tela de rodízio.
		rules, err = s.St.ListRulesActiveFrom(id, s.today())
	}
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, rules)
}

type ruleInput struct {
	DiaSemana        int    `json:"dia_semana"`
	IntervaloSemanas int    `json:"intervalo_semanas"`
	Inicio           string `json:"inicio"`
	Fim              string `json:"fim"`
	VigenteDe        string `json:"vigente_de"`
}

type replaceRulesInput struct {
	Regras    []ruleInput `json:"regras"`
	VigenteDe string      `json:"vigente_de"` // padrão: hoje
}

// replaceRules substitui o rodízio vigente do integrante: encerra as regras
// atuais na véspera de vigente_de e cria as novas — preservando o histórico
// para que meses passados continuem materializando como eram.
func (s *Server) replaceRules(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	var in replaceRulesInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "corpo inválido: "+err.Error())
		return
	}
	from := in.VigenteDe
	if from == "" {
		from = s.today()
	}
	if !validDate(from) {
		badRequest(c, "vigente_de inválida (use YYYY-MM-DD)")
		return
	}
	for _, r := range in.Regras {
		if r.DiaSemana < 0 || r.DiaSemana > 6 {
			badRequest(c, "dia_semana deve estar entre 0 (domingo) e 6 (sábado)")
			return
		}
		if !validTime(r.Inicio) || !validTime(r.Fim) {
			badRequest(c, "horários devem estar no formato HH:MM")
			return
		}
		if r.Inicio >= r.Fim {
			badRequest(c, "início deve ser antes do fim (plantões não podem virar a meia-noite)")
			return
		}
	}

	fromDate, _ := time.Parse("2006-01-02", from)
	dayBefore := fromDate.AddDate(0, 0, -1).Format("2006-01-02")

	// Encerra/remove TODAS as regras ainda relevantes a partir de from —
	// inclusive as ancoradas no futuro, que senão sobreviveriam a cada
	// "salvar" e duplicariam plantões (inflando as horas do mês).
	current, err := s.St.ListRulesActiveFrom(id, from)
	if err != nil {
		serverError(c, err)
		return
	}
	for _, r := range current {
		if r.EffectiveFrom >= from {
			if err := s.St.DeleteRule(r.ID); err != nil {
				serverError(c, err)
				return
			}
		} else if err := s.St.EndRule(r.ID, dayBefore); err != nil {
			serverError(c, err)
			return
		}
	}

	created := []store.Rule{}
	for _, r := range in.Regras {
		iv := r.IntervaloSemanas
		if iv < 1 {
			iv = 1
		}
		anchor := from
		if r.VigenteDe != "" && validDate(r.VigenteDe) {
			anchor = r.VigenteDe // âncora específica para alternância (ex.: sábado sim/não)
		}
		rule, err := s.St.CreateRule(store.Rule{
			MemberID: id, Weekday: r.DiaSemana, IntervalWeeks: iv,
			StartTime: r.Inicio, EndTime: r.Fim, EffectiveFrom: anchor,
		})
		if err != nil {
			serverError(c, err)
			return
		}
		created = append(created, rule)
	}
	s.St.Audit("rule", id, "replace", gin.H{"membro_id": id, "regras": created})
	c.JSON(http.StatusOK, created)
}

func (s *Server) deleteRule(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	rule, err := s.St.GetRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "regra não encontrada"})
		return
	}
	// Encerrar em vez de apagar preserva o histórico do calendário
	if err := s.St.EndRule(id, s.today()); err != nil {
		serverError(c, err)
		return
	}
	s.St.Audit("rule", id, "end", rule)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
