package schedule

import (
	"testing"
	"time"

	"github.com/joao/plantoes/internal/store"
)

const (
	joao    = int64(1)
	gustavo = int64(2)
)

func ptr[T any](v T) *T { return &v }

// Regras equivalentes à planilha ESCALA SUPORTE 2026:
// João 06–08 seg–sex; Gustavo 18–20 seg–sex; sábado (07–12 + 13–17)
// alternado: Gustavo nos sábados 05/09 e 19/09, João em 12/09 e 26/09.
func planilhaRules() []store.Rule {
	rules := []store.Rule{}
	for wd := 1; wd <= 5; wd++ {
		rules = append(rules,
			store.Rule{MemberID: joao, Weekday: wd, IntervalWeeks: 1, StartTime: "06:00", EndTime: "08:00", EffectiveFrom: "2026-01-01"},
			store.Rule{MemberID: gustavo, Weekday: wd, IntervalWeeks: 1, StartTime: "18:00", EndTime: "20:00", EffectiveFrom: "2026-01-01"},
		)
	}
	for _, block := range [][2]string{{"07:00", "12:00"}, {"13:00", "17:00"}} {
		rules = append(rules,
			store.Rule{MemberID: gustavo, Weekday: 6, IntervalWeeks: 2, StartTime: block[0], EndTime: block[1], EffectiveFrom: "2026-09-05"},
			store.Rule{MemberID: joao, Weekday: 6, IntervalWeeks: 2, StartTime: block[0], EndTime: block[1], EffectiveFrom: "2026-09-12"},
		)
	}
	return rules
}

func monthRange(y int, m time.Month) (time.Time, time.Time) {
	from := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	return from, from.AddDate(0, 1, -1)
}

func TestReproduzPlanilhaSetembro2026(t *testing.T) {
	from, to := monthRange(2026, time.September)
	days := Materialize(planilhaRules(), nil, from, to)
	totals := MonthlyMinutes(days)

	// 22 dias úteis × 2h = 44h + 2 sábados × 9h = 62h para cada um
	for name, id := range map[string]int64{"João": joao, "Gustavo": gustavo} {
		if got := totals[id]; got != 62*60 {
			t.Errorf("%s: esperado 62h00, obtido %s", name, FormatHours(got))
		}
	}

	// Sábado 19/09 é do Gustavo; 26/09 é do João
	check := func(date string, want int64) {
		for _, d := range days {
			if d.Date != date {
				continue
			}
			if len(d.Slots) == 0 {
				t.Fatalf("%s: sem slots", date)
			}
			for _, s := range d.Slots {
				if s.MemberID != want {
					t.Errorf("%s: slot de membro %d, esperado %d", date, s.MemberID, want)
				}
			}
		}
	}
	check("2026-09-19", gustavo)
	check("2026-09-26", joao)
}

func TestSubstituicaoConfirmada(t *testing.T) {
	from, to := monthRange(2026, time.September)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-14", Type: "substituicao",
		OriginalMemberID: ptr(joao), SubstituteMemberID: ptr(gustavo),
		StartTime: "06:00", EndTime: "08:00", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	totals := MonthlyMinutes(days)
	if got := totals[joao]; got != 60*60 {
		t.Errorf("João deveria perder 2h (60h00), obtido %s", FormatHours(got))
	}
	if got := totals[gustavo]; got != 64*60 {
		t.Errorf("Gustavo deveria ganhar 2h (64h00), obtido %s", FormatHours(got))
	}
}

func TestSubstituicaoParcialRecortaSlot(t *testing.T) {
	from := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-14", Type: "substituicao",
		OriginalMemberID: ptr(joao), SubstituteMemberID: ptr(gustavo),
		StartTime: "07:00", EndTime: "08:00", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, from, from)
	var joaoSlots, gustavoSlots []Slot
	for _, s := range days[0].Slots {
		if s.MemberID == joao {
			joaoSlots = append(joaoSlots, s)
		}
		if s.MemberID == gustavo && s.Origin == "substituicao" {
			gustavoSlots = append(gustavoSlots, s)
		}
	}
	if len(joaoSlots) != 1 || joaoSlots[0].Start != "06:00" || joaoSlots[0].End != "07:00" {
		t.Errorf("slot restante do João incorreto: %+v", joaoSlots)
	}
	if len(gustavoSlots) != 1 || gustavoSlots[0].Start != "07:00" {
		t.Errorf("slot do substituto incorreto: %+v", gustavoSlots)
	}
}

func TestTrocaAplicaNasDuasDatas(t *testing.T) {
	from, to := monthRange(2026, time.September)
	// Gustavo assume o sábado do João (26/09) e João assume o do Gustavo (19/09)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-26", Type: "troca",
		OriginalMemberID: ptr(joao), SubstituteMemberID: ptr(gustavo),
		SwapDate: ptr("2026-09-19"), StartTime: "07:00", EndTime: "17:00", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	totals := MonthlyMinutes(days)
	// Troca de sábados equivalentes: totais permanecem 62h/62h
	if totals[joao] != 62*60 || totals[gustavo] != 62*60 {
		t.Errorf("troca simétrica deveria manter 62h/62h, obtido João=%s Gustavo=%s",
			FormatHours(totals[joao]), FormatHours(totals[gustavo]))
	}
	for _, d := range days {
		if d.Date == "2026-09-19" {
			for _, s := range d.Slots {
				if s.MemberID != joao {
					t.Errorf("19/09 deveria ser do João após troca, slot: %+v", s)
				}
			}
		}
	}
}

func TestAusenciaMarcaDescoberto(t *testing.T) {
	day := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-15", Type: "ausencia",
		OriginalMemberID: ptr(joao), StartTime: "06:00", EndTime: "08:00", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, day, day)
	if !days[0].Uncovered {
		t.Error("ausência sem cobertura deveria marcar o dia como descoberto")
	}
}

func TestFeriadoZeraODia(t *testing.T) {
	from, to := monthRange(2026, time.September)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-14", Type: "feriado", Note: "Feriado municipal",
		StartTime: "00:00", EndTime: "23:59", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	totals := MonthlyMinutes(days)
	// João perde 2h (06–08) e Gustavo 2h (18–20) da segunda 14/09
	if totals[joao] != 60*60 || totals[gustavo] != 60*60 {
		t.Errorf("feriado deveria descontar as horas do dia: João=%s Gustavo=%s",
			FormatHours(totals[joao]), FormatHours(totals[gustavo]))
	}
	for _, d := range days {
		if d.Date == "2026-09-14" {
			if len(d.Slots) != 0 || !d.SemPlantao || d.Motivo != "Feriado municipal" {
				t.Errorf("dia de feriado incorreto: %+v", d)
			}
		}
	}
}

func TestFeriadoComExtraPorCima(t *testing.T) {
	day := time.Date(2026, 9, 14, 0, 0, 0, 0, time.UTC)
	exc := []store.Exception{
		{ID: 1, Date: "2026-09-14", Type: "feriado", StartTime: "00:00", EndTime: "23:59", Status: "confirmado"},
		{ID: 2, Date: "2026-09-14", Type: "extra", SubstituteMemberID: ptr(joao),
			StartTime: "08:00", EndTime: "10:00", Status: "confirmado"},
	}
	days := Materialize(planilhaRules(), exc, day, day)
	if len(days[0].Slots) != 1 || days[0].Slots[0].Origin != "extra" {
		t.Errorf("plantão individual deveria sobreviver ao feriado: %+v", days[0].Slots)
	}
}

func TestEdicaoDiaSubstituiEscala(t *testing.T) {
	from, to := monthRange(2026, time.December)
	// 24/12/2026 (quinta) vira escala de sábado: João 08–12 e 13–17
	slots := `[{"membro_id":1,"inicio":"08:00","fim":"12:00"},{"membro_id":1,"inicio":"13:00","fim":"17:00"}]`
	exc := []store.Exception{{
		ID: 1, Date: "2026-12-24", Type: "edicao_dia", SlotsJSON: &slots,
		StartTime: "00:00", EndTime: "23:59", Status: "confirmado", Note: "Véspera de Natal",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	for _, d := range days {
		if d.Date != "2026-12-24" {
			continue
		}
		if len(d.Slots) != 2 {
			t.Fatalf("dia editado deveria ter 2 slots, tem %d: %+v", len(d.Slots), d.Slots)
		}
		for _, s := range d.Slots {
			if s.MemberID != joao || s.Origin != "edicao" {
				t.Errorf("slot inesperado no dia editado: %+v", s)
			}
		}
		if d.Motivo != "Véspera de Natal" {
			t.Errorf("motivo não propagado: %q", d.Motivo)
		}
	}
	// Horas do João em 24/12: 8h em vez das 2h do rodízio (06–08)
	totals := MonthlyMinutes(days)
	base := MonthlyMinutes(Materialize(planilhaRules(), nil, from, to))
	if totals[joao]-base[joao] != 6*60 {
		t.Errorf("edição deveria somar +6h ao João, delta=%d min", totals[joao]-base[joao])
	}
	if totals[gustavo]-base[gustavo] != -2*60 {
		t.Errorf("edição deveria tirar 2h do Gustavo, delta=%d min", totals[gustavo]-base[gustavo])
	}
}

func TestFeriasSemSubstitutoRemoveEDepoisRetoma(t *testing.T) {
	from, to := monthRange(2026, time.September)
	// João de férias de 14 a 20/09 (semana com seg–sex 06–08 e o sábado 19 é do Gustavo)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-14", Type: "ferias", EndDate: ptr("2026-09-20"),
		OriginalMemberID: ptr(joao), StartTime: "00:00", EndTime: "23:59", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	totals := MonthlyMinutes(days)
	// João perde 5 dias úteis × 2h = 10h → 52h; Gustavo intacto
	if totals[joao] != 52*60 {
		t.Errorf("João deveria ficar com 52h00, obtido %s", FormatHours(totals[joao]))
	}
	if totals[gustavo] != 62*60 {
		t.Errorf("Gustavo não deveria mudar, obtido %s", FormatHours(totals[gustavo]))
	}
	for _, d := range days {
		switch d.Date {
		case "2026-09-15": // dia útil dentro das férias
			for _, s := range d.Slots {
				if s.MemberID == joao {
					t.Errorf("João não deveria ter plantão em %s: %+v", d.Date, s)
				}
			}
			if !d.Uncovered {
				t.Errorf("%s deveria estar descoberto (férias sem substituto)", d.Date)
			}
			if len(d.FeriasMemberIDs) != 1 || d.FeriasMemberIDs[0] != joao {
				t.Errorf("%s deveria listar João de férias: %v", d.Date, d.FeriasMemberIDs)
			}
		case "2026-09-21": // retorno: rodízio volta normal
			found := false
			for _, s := range d.Slots {
				if s.MemberID == joao && s.Origin == "rodizio" {
					found = true
				}
			}
			if !found {
				t.Errorf("João deveria voltar ao rodízio em %s: %+v", d.Date, d.Slots)
			}
		case "2026-09-26": // alternância do sábado preservada após as férias
			for _, s := range d.Slots {
				if s.MemberID != joao {
					t.Errorf("sábado 26/09 deveria seguir do João: %+v", s)
				}
			}
		}
	}
}

func TestFeriasComSubstitutoTransferePlantoes(t *testing.T) {
	from, to := monthRange(2026, time.September)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-14", Type: "ferias", EndDate: ptr("2026-09-20"),
		OriginalMemberID: ptr(joao), SubstituteMemberID: ptr(gustavo),
		StartTime: "00:00", EndTime: "23:59", Status: "confirmado",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	totals := MonthlyMinutes(days)
	if totals[joao] != 52*60 || totals[gustavo] != 72*60 {
		t.Errorf("esperado João=52h/Gustavo=72h, obtido João=%s Gustavo=%s",
			FormatHours(totals[joao]), FormatHours(totals[gustavo]))
	}
	for _, d := range days {
		if d.Date == "2026-09-15" {
			if d.Uncovered {
				t.Error("com substituto o dia não deveria ficar descoberto")
			}
			ok := false
			for _, s := range d.Slots {
				if s.MemberID == gustavo && s.Origin == "ferias" && s.Start == "06:00" {
					ok = true
				}
			}
			if !ok {
				t.Errorf("Gustavo deveria assumir o 06–08 de João em 15/09: %+v", d.Slots)
			}
		}
	}
}

func TestPendenteNaoAfetaHoras(t *testing.T) {
	from, to := monthRange(2026, time.September)
	exc := []store.Exception{{
		ID: 1, Date: "2026-09-14", Type: "substituicao",
		OriginalMemberID: ptr(joao), SubstituteMemberID: ptr(gustavo),
		StartTime: "06:00", EndTime: "08:00", Status: "pendente",
	}}
	days := Materialize(planilhaRules(), exc, from, to)
	totals := MonthlyMinutes(days)
	if totals[joao] != 62*60 {
		t.Errorf("pendente não deveria alterar horas, obtido %s", FormatHours(totals[joao]))
	}
	found := false
	for _, d := range days {
		if d.Date == "2026-09-14" && len(d.Pending) == 1 {
			found = true
		}
	}
	if !found {
		t.Error("exceção pendente deveria aparecer na lista de pendentes do dia")
	}
}
