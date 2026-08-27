// Package schedule materializa o calendário de plantões a partir das regras
// de rodízio e das exceções confirmadas. O calendário nunca é armazenado:
// é sempre computado sob demanda por estas funções puras.
package schedule

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/joao/plantoes/internal/store"
)

type Slot struct {
	MemberID int64  `json:"membro_id"`
	Start    string `json:"inicio"` // "HH:MM"
	End      string `json:"fim"`
	Origin   string `json:"origem"` // rodizio|substituicao|troca|extra
}

type DaySchedule struct {
	Date       string            `json:"data"` // "YYYY-MM-DD"
	Slots      []Slot            `json:"slots"`
	Pending    []store.Exception `json:"pendentes"`
	Uncovered  bool              `json:"descoberto"`
	SemPlantao bool              `json:"sem_plantao"` // dia marcado como feriado/sem plantão
	Motivo     string            `json:"motivo"`      // observação do feriado/edição do dia
	// Integrantes de férias neste dia (para a badge no calendário)
	FeriasMemberIDs []int64 `json:"ferias_membro_ids"`
}

// EditedSlot é o formato dos slots gravados em slots_json de uma exceção
// do tipo edicao_dia.
type EditedSlot struct {
	MemberID int64  `json:"membro_id"`
	Start    string `json:"inicio"`
	End      string `json:"fim"`
}

const dateLayout = "2006-01-02"

// Materialize computa a escala de cada dia em [from, to] (inclusive).
func Materialize(rules []store.Rule, exceptions []store.Exception, from, to time.Time) []DaySchedule {
	days := []DaySchedule{}
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		days = append(days, materializeDay(rules, exceptions, d))
	}
	return days
}

func materializeDay(rules []store.Rule, exceptions []store.Exception, day time.Time) DaySchedule {
	date := day.Format(dateLayout)
	weekday := int(day.Weekday()) // 0=domingo, como na tabela

	ds := DaySchedule{Date: date, Slots: []Slot{}, Pending: []store.Exception{}}

	// 1. Slots base do rodízio
	for _, r := range rules {
		if r.Weekday != weekday {
			continue
		}
		if r.EffectiveFrom > date {
			continue
		}
		if r.EffectiveTo != nil && *r.EffectiveTo < date {
			continue
		}
		if !weekMatches(r, day) {
			continue
		}
		ds.Slots = append(ds.Slots, Slot{MemberID: r.MemberID, Start: r.StartTime, End: r.EndTime, Origin: "rodizio"})
	}

	// 2. Férias confirmadas mascaram o rodízio do integrante no período
	// [date, end_date]. As regras continuam correndo por baixo, então no
	// retorno a escala (inclusive alternâncias) volta exatamente como era.
	for _, e := range exceptions {
		if e.Status != "confirmado" || e.Type != "ferias" || e.EndDate == nil || e.OriginalMemberID == nil {
			continue
		}
		if date < e.Date || date > *e.EndDate {
			continue
		}
		ds.FeriasMemberIDs = append(ds.FeriasMemberIDs, *e.OriginalMemberID)
		if e.SubstituteMemberID != nil {
			ds.Slots = transferInterval(ds.Slots, e.OriginalMemberID, e.SubstituteMemberID, "00:00", "24:00", "ferias")
		} else {
			had := false
			for _, s := range ds.Slots {
				if s.MemberID == *e.OriginalMemberID {
					had = true
				}
			}
			ds.Slots = removeInterval(ds.Slots, e.OriginalMemberID, "00:00", "24:00")
			if had {
				ds.Uncovered = true
			}
		}
	}

	// 3. Exceções de dia inteiro aplicam em seguida: feriado zera a escala
	// automática do dia; edicao_dia substitui a escala inteira pelos slots
	// definidos pelo gestor. Exceções individuais (extra etc.) confirmadas
	// depois ainda se aplicam por cima.
	for _, e := range exceptions {
		if e.Status != "confirmado" || e.Date != date {
			continue
		}
		switch e.Type {
		case "feriado":
			ds.Slots = []Slot{}
			ds.SemPlantao = true
			ds.Motivo = e.Note
		case "edicao_dia":
			ds.Slots = []Slot{}
			ds.Motivo = e.Note
			if e.SlotsJSON != nil {
				var edited []EditedSlot
				if json.Unmarshal([]byte(*e.SlotsJSON), &edited) == nil {
					for _, s := range edited {
						ds.Slots = append(ds.Slots, Slot{MemberID: s.MemberID, Start: s.Start, End: s.End, Origin: "edicao"})
					}
				}
			}
		}
	}

	// 4. Exceções individuais confirmadas sobrepõem o resultado
	for _, e := range exceptions {
		if e.Status != "confirmado" {
			if e.Status == "pendente" && (e.Date == date || (e.SwapDate != nil && *e.SwapDate == date)) {
				ds.Pending = append(ds.Pending, e)
			}
			continue
		}
		switch {
		case (e.Type == "substituicao" || e.Type == "troca") && e.Date == date:
			// O substituto assume os trechos do original dentro da faixa,
			// preservando a estrutura dos slots (ex.: intervalo de almoço)
			ds.Slots = transferInterval(ds.Slots, e.OriginalMemberID, e.SubstituteMemberID, e.StartTime, e.EndTime, e.Type)
		case e.Type == "troca" && e.SwapDate != nil && *e.SwapDate == date:
			// Na data de retorno, o original assume os plantões do substituto
			if e.OriginalMemberID != nil && e.SubstituteMemberID != nil {
				for i := range ds.Slots {
					if ds.Slots[i].MemberID == *e.SubstituteMemberID {
						ds.Slots[i].MemberID = *e.OriginalMemberID
						ds.Slots[i].Origin = "troca"
					}
				}
			}
		case e.Type == "ausencia" && e.Date == date:
			ds.Slots = removeInterval(ds.Slots, e.OriginalMemberID, e.StartTime, e.EndTime)
			if !covered(ds.Slots, e.StartTime, e.EndTime) {
				ds.Uncovered = true
			}
		case e.Type == "extra" && e.Date == date:
			if e.SubstituteMemberID != nil {
				ds.Slots = append(ds.Slots, Slot{MemberID: *e.SubstituteMemberID, Start: e.StartTime, End: e.EndTime, Origin: "extra"})
			}
		}
	}

	sort.Slice(ds.Slots, func(i, j int) bool {
		if ds.Slots[i].Start != ds.Slots[j].Start {
			return ds.Slots[i].Start < ds.Slots[j].Start
		}
		return ds.Slots[i].MemberID < ds.Slots[j].MemberID
	})
	return ds
}

// weekMatches trata regras com intervalo de semanas > 1 (ex.: sábado
// alternado): conta semanas a partir da primeira ocorrência do dia da
// semana em ou após effective_from e exige múltiplo do intervalo.
func weekMatches(r store.Rule, day time.Time) bool {
	if r.IntervalWeeks <= 1 {
		return true
	}
	anchor, err := time.Parse(dateLayout, r.EffectiveFrom)
	if err != nil {
		return true
	}
	// primeira ocorrência do weekday da regra em ou após effective_from
	offset := (r.Weekday - int(anchor.Weekday()) + 7) % 7
	first := anchor.AddDate(0, 0, offset)
	weeks := int(day.Sub(first).Hours() / (24 * 7))
	return weeks%r.IntervalWeeks == 0
}

// transferInterval reatribui ao substituto os trechos dos slots do membro
// original que intersectam [start, end), mantendo as partes de fora com o
// original.
func transferInterval(slots []Slot, fromID, toID *int64, start, end, origin string) []Slot {
	if fromID == nil || toID == nil {
		return slots
	}
	out := []Slot{}
	for _, s := range slots {
		if s.MemberID != *fromID || s.End <= start || s.Start >= end {
			out = append(out, s)
			continue
		}
		if s.Start < start {
			left := s
			left.End = start
			out = append(out, left)
		}
		mid := s
		mid.MemberID = *toID
		mid.Origin = origin
		if mid.Start < start {
			mid.Start = start
		}
		if mid.End > end {
			mid.End = end
		}
		out = append(out, mid)
		if s.End > end {
			right := s
			right.Start = end
			out = append(out, right)
		}
	}
	return out
}

// removeInterval recorta [start, end) dos slots do membro indicado,
// preservando as partes que não se sobrepõem.
func removeInterval(slots []Slot, memberID *int64, start, end string) []Slot {
	if memberID == nil {
		return slots
	}
	out := []Slot{}
	for _, s := range slots {
		if s.MemberID != *memberID || s.End <= start || s.Start >= end {
			out = append(out, s)
			continue
		}
		if s.Start < start {
			left := s
			left.End = start
			out = append(out, left)
		}
		if s.End > end {
			right := s
			right.Start = end
			out = append(out, right)
		}
	}
	return out
}

// covered verifica se [start, end) está totalmente coberto pela união dos slots.
func covered(slots []Slot, start, end string) bool {
	type iv struct{ s, e int }
	ivs := []iv{}
	for _, s := range slots {
		ivs = append(ivs, iv{toMin(s.Start), toMin(s.End)})
	}
	sort.Slice(ivs, func(i, j int) bool { return ivs[i].s < ivs[j].s })
	cur := toMin(start)
	target := toMin(end)
	for _, v := range ivs {
		if v.s > cur {
			break
		}
		if v.e > cur {
			cur = v.e
		}
		if cur >= target {
			return true
		}
	}
	return cur >= target
}

func toMin(hhmm string) int {
	if len(hhmm) < 5 {
		return 0
	}
	h := int(hhmm[0]-'0')*10 + int(hhmm[1]-'0')
	m := int(hhmm[3]-'0')*10 + int(hhmm[4]-'0')
	return h*60 + m
}
