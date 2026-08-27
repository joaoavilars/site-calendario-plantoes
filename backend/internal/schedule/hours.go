package schedule

import "fmt"

// MonthlyMinutes soma os minutos de plantão de cada membro nos dias materializados.
func MonthlyMinutes(days []DaySchedule) map[int64]int {
	totals := map[int64]int{}
	for _, d := range days {
		for _, s := range d.Slots {
			totals[s.MemberID] += toMin(s.End) - toMin(s.Start)
		}
	}
	return totals
}

// FormatHours converte minutos em "62h00".
func FormatHours(minutes int) string {
	return fmt.Sprintf("%dh%02d", minutes/60, minutes%60)
}
