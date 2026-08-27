package notify

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/joao/plantoes/internal/config"
	"github.com/joao/plantoes/internal/schedule"
	"github.com/joao/plantoes/internal/store"
)

type Scheduler struct {
	cfg  *config.Config
	st   *store.Store
	tg   *Telegram
	cron *cron.Cron
}

func NewScheduler(cfg *config.Config, st *store.Store, tg *Telegram) *Scheduler {
	return &Scheduler{
		cfg:  cfg,
		st:   st,
		tg:   tg,
		cron: cron.New(cron.WithLocation(cfg.Location)),
	}
}

func (s *Scheduler) Start() error {
	if s.cfg.Alertas.ResumoDiario.Habilitado {
		var h, m int
		if _, err := fmt.Sscanf(s.cfg.Alertas.ResumoDiario.Horario, "%d:%d", &h, &m); err != nil {
			return fmt.Errorf("horário do resumo diário inválido %q: %w", s.cfg.Alertas.ResumoDiario.Horario, err)
		}
		spec := fmt.Sprintf("%d %d * * *", m, h)
		if _, err := s.cron.AddFunc(spec, s.sendDailySummary); err != nil {
			return err
		}
	}
	if s.cfg.Alertas.LembretePlantao.Habilitado {
		if _, err := s.cron.AddFunc("* * * * *", s.sendReminders); err != nil {
			return err
		}
	}
	s.cron.Start()
	return nil
}

func (s *Scheduler) Stop() { s.cron.Stop() }

// materializeDay computa a escala do dia com nomes de membros resolvidos.
func (s *Scheduler) materializeDay(day time.Time) (schedule.DaySchedule, map[int64]store.Member, error) {
	date := day.Format("2006-01-02")
	rules, err := s.st.ListRulesForRange(date, date)
	if err != nil {
		return schedule.DaySchedule{}, nil, err
	}
	excs, err := s.st.ListExceptionsForRange(date, date, []string{"confirmado"})
	if err != nil {
		return schedule.DaySchedule{}, nil, err
	}
	days := schedule.Materialize(rules, excs, day, day)
	members, err := s.st.ListMembers(true)
	if err != nil {
		return schedule.DaySchedule{}, nil, err
	}
	byID := map[int64]store.Member{}
	for _, m := range members {
		byID[m.ID] = m
	}
	return days[0], byID, nil
}

func memberLabel(m store.Member) string {
	if m.TelegramMention != "" {
		return fmt.Sprintf("%s (%s)", m.Name, m.TelegramMention)
	}
	return m.Name
}

func (s *Scheduler) sendDailySummary() {
	now := time.Now().In(s.cfg.Location)
	key := "resumo:" + now.Format("2006-01-02")
	fresh, err := s.st.MarkAlertSent(key)
	if err != nil || !fresh {
		return
	}
	day, members, err := s.materializeDay(now)
	if err != nil {
		log.Printf("[alertas] resumo diário: %v", err)
		return
	}
	var b strings.Builder
	fmt.Fprintf(&b, "📅 Escala de plantões de hoje (%s):\n", now.Format("02/01/2006"))
	if len(day.Slots) == 0 {
		b.WriteString("Sem plantões individuais hoje.")
	}
	for _, slot := range day.Slots {
		fmt.Fprintf(&b, "• %s às %s — %s\n", slot.Start, slot.End, memberLabel(members[slot.MemberID]))
	}
	for _, id := range day.FeriasMemberIDs {
		fmt.Fprintf(&b, "🏖️ %s está de férias\n", members[id].Name)
	}
	if day.Uncovered {
		b.WriteString("⚠️ Atenção: há horário descoberto hoje!")
	}
	if err := s.tg.Send(b.String()); err != nil {
		log.Printf("[alertas] resumo diário: %v", err)
	}
}

func (s *Scheduler) sendReminders() {
	now := time.Now().In(s.cfg.Location)
	lead := time.Duration(s.cfg.Alertas.LembretePlantao.AntecedenciaMinutos) * time.Minute
	target := now.Add(lead)
	day, members, err := s.materializeDay(target)
	if err != nil {
		log.Printf("[alertas] lembrete: %v", err)
		return
	}
	targetHM := target.Format("15:04")
	for _, slot := range day.Slots {
		if slot.Start != targetHM {
			continue
		}
		key := fmt.Sprintf("lembrete:%s:%d:%s", target.Format("2006-01-02"), slot.MemberID, slot.Start)
		fresh, err := s.st.MarkAlertSent(key)
		if err != nil || !fresh {
			continue
		}
		msg := fmt.Sprintf("⏰ Lembrete: plantão de %s começa às %s (em %d min).",
			memberLabel(members[slot.MemberID]), slot.Start, s.cfg.Alertas.LembretePlantao.AntecedenciaMinutos)
		if err := s.tg.Send(msg); err != nil {
			log.Printf("[alertas] lembrete: %v", err)
		}
	}
}

// NotifyExceptionConfirmed é chamado pelo handler de confirmação (alerta tipo c).
func (s *Scheduler) NotifyExceptionConfirmed(e store.Exception) {
	if !s.cfg.Alertas.AvisoTrocaConfirmada.Habilitado {
		return
	}
	members, err := s.st.ListMembers(true)
	if err != nil {
		log.Printf("[alertas] aviso de troca: %v", err)
		return
	}
	byID := map[int64]store.Member{}
	for _, m := range members {
		byID[m.ID] = m
	}
	name := func(id *int64) string {
		if id == nil {
			return "?"
		}
		return memberLabel(byID[*id])
	}
	dateBR := func(iso string) string {
		if t, err := time.Parse("2006-01-02", iso); err == nil {
			return t.Format("02/01/2006")
		}
		return iso
	}

	var msg string
	switch e.Type {
	case "substituicao":
		msg = fmt.Sprintf("🔁 Substituição confirmada: %s assume o plantão de %s em %s, das %s às %s.",
			name(e.SubstituteMemberID), name(e.OriginalMemberID), dateBR(e.Date), e.StartTime, e.EndTime)
	case "troca":
		swap := ""
		if e.SwapDate != nil {
			swap = fmt.Sprintf(" Em %s, %s assume os plantões de %s.",
				dateBR(*e.SwapDate), name(e.OriginalMemberID), name(e.SubstituteMemberID))
		}
		msg = fmt.Sprintf("🔀 Troca confirmada: %s assume o plantão de %s em %s, das %s às %s.%s",
			name(e.SubstituteMemberID), name(e.OriginalMemberID), dateBR(e.Date), e.StartTime, e.EndTime, swap)
	case "extra":
		msg = fmt.Sprintf("➕ Plantão extra confirmado: %s em %s, das %s às %s.",
			name(e.SubstituteMemberID), dateBR(e.Date), e.StartTime, e.EndTime)
	case "ausencia":
		msg = fmt.Sprintf("🚫 Ausência confirmada: %s não fará o plantão de %s, das %s às %s.",
			name(e.OriginalMemberID), dateBR(e.Date), e.StartTime, e.EndTime)
	case "feriado":
		msg = fmt.Sprintf("🎉 Dia sem plantão confirmado: %s.", dateBR(e.Date))
	case "ferias":
		fim := e.Date
		if e.EndDate != nil {
			fim = *e.EndDate
		}
		msg = fmt.Sprintf("🏖️ Férias confirmadas: %s de %s a %s.",
			name(e.OriginalMemberID), dateBR(e.Date), dateBR(fim))
		if e.SubstituteMemberID != nil {
			msg += fmt.Sprintf(" %s cobre os plantões no período.", name(e.SubstituteMemberID))
		} else {
			msg += " Sem substituto definido — plantões do período ficam descobertos."
		}
	case "edicao_dia":
		detalhe := ""
		if e.SlotsJSON != nil {
			var edited []schedule.EditedSlot
			if json.Unmarshal([]byte(*e.SlotsJSON), &edited) == nil {
				partes := []string{}
				for _, s := range edited {
					partes = append(partes, fmt.Sprintf("%s %s–%s", memberLabel(byID[s.MemberID]), s.Start, s.End))
				}
				detalhe = " Nova escala: " + strings.Join(partes, ", ") + "."
			}
		}
		msg = fmt.Sprintf("🛠️ Escala do dia %s ajustada manualmente.%s", dateBR(e.Date), detalhe)
	}
	if e.Note != "" {
		msg += " Obs: " + e.Note
	}
	if err := s.tg.Send(msg); err != nil {
		log.Printf("[alertas] aviso de confirmação: %v", err)
	}
}
