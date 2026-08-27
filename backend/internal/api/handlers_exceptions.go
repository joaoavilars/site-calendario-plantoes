package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/store"
)

type slotInput struct {
	MembroID int64  `json:"membro_id"`
	Inicio   string `json:"inicio"`
	Fim      string `json:"fim"`
}

type exceptionInput struct {
	Data               string      `json:"data"`
	Tipo               string      `json:"tipo"`
	MembroOriginalID   *int64      `json:"membro_original_id"`
	MembroSubstitutoID *int64      `json:"membro_substituto_id"`
	DataTroca          *string     `json:"data_troca"`
	DataFim            *string     `json:"data_fim"` // ferias: último dia do período
	Inicio             string      `json:"inicio"`
	Fim                string      `json:"fim"`
	Slots              []slotInput `json:"slots"` // edicao_dia: nova escala completa do dia
	Observacao         string      `json:"observacao"`
}

func (in exceptionInput) isDayLevel() bool {
	return in.Tipo == "feriado" || in.Tipo == "edicao_dia" || in.Tipo == "ferias"
}

type previewDay struct {
	Data   string     `json:"data"`
	Antes  []slotView `json:"antes"`
	Depois []slotView `json:"depois"`
}

type previewView struct {
	Dias        []previewDay `json:"dias"`
	HorasAntes  []hoursView  `json:"horas_antes"`
	HorasDepois []hoursView  `json:"horas_depois"`
}

func (in exceptionInput) validate() string {
	if !validDate(in.Data) {
		return "data inválida (use YYYY-MM-DD)"
	}
	if !in.isDayLevel() {
		if !validTime(in.Inicio) || !validTime(in.Fim) {
			return "horários devem estar no formato HH:MM"
		}
		if in.Inicio >= in.Fim {
			return "início deve ser antes do fim"
		}
	}
	switch in.Tipo {
	case "feriado":
		return ""
	case "ferias":
		if in.MembroOriginalID == nil {
			return "férias exigem membro_original_id (quem sai de férias)"
		}
		if in.DataFim == nil || !validDate(*in.DataFim) {
			return "férias exigem data_fim válida (YYYY-MM-DD)"
		}
		if *in.DataFim < in.Data {
			return "data_fim deve ser igual ou posterior à data de início"
		}
		ini, _ := time.Parse("2006-01-02", in.Data)
		fim, _ := time.Parse("2006-01-02", *in.DataFim)
		if fim.Sub(ini) > 90*24*time.Hour {
			return "período de férias não pode passar de 90 dias"
		}
		if in.MembroSubstitutoID != nil && *in.MembroSubstitutoID == *in.MembroOriginalID {
			return "o substituto deve ser diferente de quem sai de férias"
		}
		return ""
	case "edicao_dia":
		if len(in.Slots) == 0 {
			return "edição do dia exige ao menos um plantão em slots (use 'feriado' para dia sem plantão)"
		}
		for _, s := range in.Slots {
			if s.MembroID == 0 {
				return "cada plantão da edição precisa de membro_id"
			}
			if !validTime(s.Inicio) || !validTime(s.Fim) || s.Inicio >= s.Fim {
				return "horários dos plantões da edição devem ser HH:MM com início antes do fim"
			}
		}
		return ""
	}
	switch in.Tipo {
	case "substituicao", "troca":
		if in.MembroOriginalID == nil || in.MembroSubstitutoID == nil {
			return "substituição/troca exige membro_original_id e membro_substituto_id"
		}
		if *in.MembroOriginalID == *in.MembroSubstitutoID {
			return "os dois integrantes devem ser diferentes"
		}
		if in.Tipo == "troca" && (in.DataTroca == nil || !validDate(*in.DataTroca)) {
			return "troca exige data_troca válida (YYYY-MM-DD)"
		}
	case "extra":
		if in.MembroSubstitutoID == nil {
			return "plantão extra exige membro_substituto_id"
		}
	case "ausencia":
		if in.MembroOriginalID == nil {
			return "ausência exige membro_original_id"
		}
	default:
		return "tipo deve ser substituicao, troca, extra ou ausencia"
	}
	return ""
}

func slotViewsEqual(a, b []slotView) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// buildPreview compara antes/depois nos meses e dias afetados pela exceção.
func (s *Server) buildPreview(e store.Exception) (previewView, error) {
	confirmed := e
	confirmed.Status = "confirmado"

	dates := []string{e.Date}
	if e.SwapDate != nil && *e.SwapDate != e.Date {
		dates = append(dates, *e.SwapDate)
	}
	if e.Type == "ferias" && e.EndDate != nil {
		dates = nil
		ini, err1 := time.Parse("2006-01-02", e.Date)
		fim, err2 := time.Parse("2006-01-02", *e.EndDate)
		if err1 == nil && err2 == nil {
			for d := ini; !d.After(fim); d = d.AddDate(0, 0, 1) {
				dates = append(dates, d.Format("2006-01-02"))
			}
		}
	}

	members, err := s.membersByID()
	if err != nil {
		return previewView{}, err
	}

	pv := previewView{Dias: []previewDay{}}
	monthsDone := map[string]bool{}
	for i, ds := range dates {
		t, err := time.Parse("2006-01-02", ds)
		if err != nil {
			continue
		}
		monthKey := t.Format("2006-01")
		if monthsDone[monthKey] {
			continue
		}
		monthsDone[monthKey] = true

		before, err := s.materializeMonth(t.Year(), int(t.Month()), nil)
		if err != nil {
			return previewView{}, err
		}
		after, err := s.materializeMonth(t.Year(), int(t.Month()), &confirmed)
		if err != nil {
			return previewView{}, err
		}
		for _, target := range dates {
			for j := range before {
				if before[j].Date != target {
					continue
				}
				antes := enrichDays(before[j:j+1], members)[0].Slots
				depois := enrichDays(after[j:j+1], members)[0].Slots
				// Em períodos longos (férias), mostra só os dias que mudam
				if e.Type == "ferias" && slotViewsEqual(antes, depois) {
					continue
				}
				if len(pv.Dias) < 20 {
					pv.Dias = append(pv.Dias, previewDay{Data: target, Antes: antes, Depois: depois})
				}
			}
		}
		if i == 0 {
			pv.HorasAntes = hoursSummary(before, members)
			pv.HorasDepois = hoursSummary(after, members)
		}
	}
	return pv, nil
}

func (s *Server) createException(c *gin.Context) {
	var in exceptionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "corpo inválido: "+err.Error())
		return
	}
	if msg := in.validate(); msg != "" {
		badRequest(c, msg)
		return
	}
	var slotsJSON *string
	if in.isDayLevel() {
		in.Inicio, in.Fim = "00:00", "23:59"
		in.DataTroca = nil
		if in.Tipo != "ferias" {
			in.MembroOriginalID, in.MembroSubstitutoID, in.DataFim = nil, nil, nil
		}
		if in.Tipo == "edicao_dia" {
			raw, err := json.Marshal(in.Slots)
			if err != nil {
				badRequest(c, "slots inválidos")
				return
			}
			str := string(raw)
			slotsJSON = &str
		}
	}
	e, err := s.St.CreateException(store.Exception{
		Date: in.Data, Type: in.Tipo,
		OriginalMemberID: in.MembroOriginalID, SubstituteMemberID: in.MembroSubstitutoID,
		SwapDate: in.DataTroca, EndDate: in.DataFim, StartTime: in.Inicio, EndTime: in.Fim,
		SlotsJSON: slotsJSON, Note: in.Observacao,
	})
	if err != nil {
		serverError(c, err)
		return
	}
	s.St.Audit("exception", e.ID, "create", e)
	preview, err := s.buildPreview(e)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"excecao": e, "preview": preview})
}

func (s *Server) listExceptions(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("ano", "0"))
	month, _ := strconv.Atoi(c.DefaultQuery("mes", "0"))
	status := c.Query("status")

	from, to := "0000-01-01", "9999-12-31"
	if year > 0 && month >= 1 && month <= 12 {
		f := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		from, to = f.Format("2006-01-02"), f.AddDate(0, 1, -1).Format("2006-01-02")
	}
	statuses := []string{}
	if status != "" {
		statuses = append(statuses, status)
	}
	excs, err := s.St.ListExceptionsForRange(from, to, statuses)
	if err != nil {
		serverError(c, err)
		return
	}
	// filtros opcionais: tipo e membro (original OU substituto)
	tipo := c.Query("tipo")
	membro, _ := strconv.ParseInt(c.DefaultQuery("membro", "0"), 10, 64)
	if tipo != "" || membro > 0 {
		filtered := []store.Exception{}
		for _, e := range excs {
			if tipo != "" && e.Type != tipo {
				continue
			}
			if membro > 0 {
				isOrig := e.OriginalMemberID != nil && *e.OriginalMemberID == membro
				isSub := e.SubstituteMemberID != nil && *e.SubstituteMemberID == membro
				if !isOrig && !isSub {
					continue
				}
			}
			filtered = append(filtered, e)
		}
		excs = filtered
	}
	c.JSON(http.StatusOK, excs)
}

func (s *Server) setExceptionStatus(c *gin.Context, from []string, to string) (store.Exception, bool) {
	id, ok := paramID(c)
	if !ok {
		return store.Exception{}, false
	}
	e, err := s.St.GetException(id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"erro": "agendamento não encontrado"})
		return store.Exception{}, false
	}
	if err != nil {
		serverError(c, err)
		return store.Exception{}, false
	}
	allowed := false
	for _, st := range from {
		if e.Status == st {
			allowed = true
		}
	}
	if !allowed {
		badRequest(c, "status atual ("+e.Status+") não permite esta operação")
		return store.Exception{}, false
	}
	e, err = s.St.SetExceptionStatus(id, to)
	if err != nil {
		serverError(c, err)
		return store.Exception{}, false
	}
	s.St.Audit("exception", e.ID, to, e)
	return e, true
}

func (s *Server) confirmException(c *gin.Context) {
	e, ok := s.setExceptionStatus(c, []string{"pendente"}, "confirmado")
	if !ok {
		return
	}
	if s.Sched != nil {
		go s.Sched.NotifyExceptionConfirmed(e)
	}
	c.JSON(http.StatusOK, e)
}

func (s *Server) rejectException(c *gin.Context) {
	e, ok := s.setExceptionStatus(c, []string{"pendente"}, "rejeitado")
	if !ok {
		return
	}
	c.JSON(http.StatusOK, e)
}

func (s *Server) cancelException(c *gin.Context) {
	e, ok := s.setExceptionStatus(c, []string{"pendente", "confirmado"}, "cancelado")
	if !ok {
		return
	}
	c.JSON(http.StatusOK, e)
}

func (s *Server) listAudit(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limite", "200"))
	entries, err := s.St.ListAudit(c.Query("entidade"), c.Query("de"), c.Query("ate"), limit)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (s *Server) testAlert(c *gin.Context) {
	if !s.Tg.Enabled() {
		badRequest(c, "telegram não configurado (bot_token/chat_id ausentes)")
		return
	}
	if err := s.Tg.Send("✅ Teste de alerta do Calendário de Plantões — configuração OK!"); err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
