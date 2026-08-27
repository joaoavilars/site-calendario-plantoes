package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/schedule"
	"github.com/joao/plantoes/internal/store"
)

type slotView struct {
	MembroID int64  `json:"membro_id"`
	Nome     string `json:"nome"`
	Cor      string `json:"cor"`
	Inicio   string `json:"inicio"`
	Fim      string `json:"fim"`
	Origem   string `json:"origem"`
}

type dayView struct {
	Data       string            `json:"data"`
	Slots      []slotView        `json:"slots"`
	Pendentes  []store.Exception `json:"pendentes"`
	Descoberto bool              `json:"descoberto"`
	SemPlantao bool              `json:"sem_plantao"`
	Motivo     string            `json:"motivo"`
	Ferias     []feriasView      `json:"ferias"`
}

type feriasView struct {
	MembroID int64  `json:"membro_id"`
	Nome     string `json:"nome"`
}

type hoursView struct {
	MembroID       int64  `json:"membro_id"`
	Nome           string `json:"nome"`
	Cor            string `json:"cor"`
	TotalMinutos   int    `json:"total_minutos"`
	TotalFormatado string `json:"total_formatado"`
}

type calendarView struct {
	Ano   int         `json:"ano"`
	Mes   int         `json:"mes"`
	Dias  []dayView   `json:"dias"`
	Horas []hoursView `json:"horas"`
}

func (s *Server) membersByID() (map[int64]store.Member, error) {
	members, err := s.St.ListMembers(true)
	if err != nil {
		return nil, err
	}
	byID := map[int64]store.Member{}
	for _, m := range members {
		byID[m.ID] = m
	}
	return byID, nil
}

func enrichDays(days []schedule.DaySchedule, members map[int64]store.Member) []dayView {
	out := []dayView{}
	for _, d := range days {
		dv := dayView{
			Data: d.Date, Slots: []slotView{}, Pendentes: d.Pending,
			Descoberto: d.Uncovered, SemPlantao: d.SemPlantao, Motivo: d.Motivo,
			Ferias: []feriasView{},
		}
		for _, id := range d.FeriasMemberIDs {
			dv.Ferias = append(dv.Ferias, feriasView{MembroID: id, Nome: members[id].Name})
		}
		for _, sl := range d.Slots {
			m := members[sl.MemberID]
			dv.Slots = append(dv.Slots, slotView{
				MembroID: sl.MemberID, Nome: m.Name, Cor: m.Color,
				Inicio: sl.Start, Fim: sl.End, Origem: sl.Origin,
			})
		}
		out = append(out, dv)
	}
	return out
}

func hoursSummary(days []schedule.DaySchedule, members map[int64]store.Member) []hoursView {
	totals := schedule.MonthlyMinutes(days)
	out := []hoursView{}
	// ordena por nome para exibição estável
	for id, m := range members {
		if minutes, ok := totals[id]; ok {
			out = append(out, hoursView{
				MembroID: id, Nome: m.Name, Cor: m.Color,
				TotalMinutos: minutes, TotalFormatado: schedule.FormatHours(minutes),
			})
		} else if m.Active {
			out = append(out, hoursView{MembroID: id, Nome: m.Name, Cor: m.Color, TotalFormatado: "0h00"})
		}
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Nome < out[i].Nome {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

// materializeMonth computa o mês com todas as exceções relevantes do banco,
// opcionalmente tratando uma exceção extra como se estivesse confirmada
// (usado no preview de confirmação).
func (s *Server) materializeMonth(year, month int, hypothetical *store.Exception) ([]schedule.DaySchedule, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)
	fromS, toS := from.Format("2006-01-02"), to.Format("2006-01-02")

	rules, err := s.St.ListRulesForRange(fromS, toS)
	if err != nil {
		return nil, err
	}
	excs, err := s.St.ListExceptionsForRange(fromS, toS, []string{"confirmado", "pendente"})
	if err != nil {
		return nil, err
	}
	if hypothetical != nil {
		found := false
		for i := range excs {
			if excs[i].ID == hypothetical.ID {
				excs[i] = *hypothetical
				found = true
			}
		}
		if !found {
			excs = append(excs, *hypothetical)
		}
	}
	return schedule.Materialize(rules, excs, from, to), nil
}

func (s *Server) getCalendar(c *gin.Context) {
	year, err1 := strconv.Atoi(c.Param("ano"))
	month, err2 := strconv.Atoi(c.Param("mes"))
	if err1 != nil || err2 != nil || month < 1 || month > 12 || year < 2000 || year > 2200 {
		badRequest(c, "ano/mês inválidos")
		return
	}
	days, err := s.materializeMonth(year, month, nil)
	if err != nil {
		serverError(c, err)
		return
	}
	members, err := s.membersByID()
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, calendarView{
		Ano: year, Mes: month,
		Dias:  enrichDays(days, members),
		Horas: hoursSummary(days, members),
	})
}
