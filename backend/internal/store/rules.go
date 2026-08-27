package store

type Rule struct {
	ID            int64   `json:"id"`
	MemberID      int64   `json:"membro_id"`
	Weekday       int     `json:"dia_semana"`      // 0=domingo ... 6=sábado
	IntervalWeeks int     `json:"intervalo_semanas"` // 1=toda semana, 2=semana sim/semana não...
	StartTime     string  `json:"inicio"`          // "HH:MM"
	EndTime       string  `json:"fim"`
	EffectiveFrom string  `json:"vigente_de"` // "YYYY-MM-DD"; para intervalo>1, âncora da alternância
	EffectiveTo   *string `json:"vigente_ate"`
}

const ruleCols = "id, member_id, weekday, interval_weeks, start_time, end_time, effective_from, effective_to"

func scanRule(row interface{ Scan(...any) error }) (Rule, error) {
	var r Rule
	err := row.Scan(&r.ID, &r.MemberID, &r.Weekday, &r.IntervalWeeks, &r.StartTime, &r.EndTime, &r.EffectiveFrom, &r.EffectiveTo)
	return r, err
}

func (s *Store) ListRulesByMember(memberID int64, onlyCurrent bool, today string) ([]Rule, error) {
	q := "SELECT " + ruleCols + " FROM rotation_rules WHERE member_id = ?"
	args := []any{memberID}
	if onlyCurrent {
		q += " AND effective_from <= ? AND (effective_to IS NULL OR effective_to >= ?)"
		args = append(args, today, today)
	}
	q += " ORDER BY weekday, start_time"
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Rule{}
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListRulesActiveFrom retorna todas as regras do membro ainda relevantes a
// partir de from — inclusive as com effective_from no futuro (ex.: âncora de
// alternância), que um filtro de "vigente hoje" deixaria escapar.
func (s *Store) ListRulesActiveFrom(memberID int64, from string) ([]Rule, error) {
	rows, err := s.DB.Query(
		"SELECT "+ruleCols+" FROM rotation_rules WHERE member_id = ? AND (effective_to IS NULL OR effective_to >= ?) ORDER BY weekday, start_time",
		memberID, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Rule{}
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListRulesForRange retorna regras cuja vigência intersecta [from, to].
func (s *Store) ListRulesForRange(from, to string) ([]Rule, error) {
	rows, err := s.DB.Query(
		"SELECT "+ruleCols+" FROM rotation_rules WHERE effective_from <= ? AND (effective_to IS NULL OR effective_to >= ?)",
		to, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Rule{}
	for rows.Next() {
		r, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetRule(id int64) (Rule, error) {
	row := s.DB.QueryRow("SELECT "+ruleCols+" FROM rotation_rules WHERE id = ?", id)
	return scanRule(row)
}

func (s *Store) CreateRule(r Rule) (Rule, error) {
	if r.IntervalWeeks < 1 {
		r.IntervalWeeks = 1
	}
	res, err := s.DB.Exec(
		"INSERT INTO rotation_rules (member_id, weekday, interval_weeks, start_time, end_time, effective_from, effective_to) VALUES (?, ?, ?, ?, ?, ?, ?)",
		r.MemberID, r.Weekday, r.IntervalWeeks, r.StartTime, r.EndTime, r.EffectiveFrom, r.EffectiveTo)
	if err != nil {
		return Rule{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetRule(id)
}

// EndRule encerra a vigência de uma regra em endDate (exclusivo a partir do dia seguinte).
func (s *Store) EndRule(id int64, endDate string) error {
	_, err := s.DB.Exec("UPDATE rotation_rules SET effective_to = ? WHERE id = ?", endDate, id)
	return err
}

func (s *Store) DeleteRule(id int64) error {
	_, err := s.DB.Exec("DELETE FROM rotation_rules WHERE id = ?", id)
	return err
}
