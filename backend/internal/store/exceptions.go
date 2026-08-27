package store

type Exception struct {
	ID                 int64   `json:"id"`
	Date               string  `json:"data"` // "YYYY-MM-DD"
	Type               string  `json:"tipo"` // substituicao|troca|extra|ausencia|feriado|edicao_dia
	OriginalMemberID   *int64  `json:"membro_original_id"`
	SubstituteMemberID *int64  `json:"membro_substituto_id"`
	SwapDate           *string `json:"data_troca"`
	EndDate            *string `json:"data_fim"` // ferias: último dia do período
	StartTime          string  `json:"inicio"`
	EndTime            string  `json:"fim"`
	SlotsJSON          *string `json:"slots_json"` // edicao_dia: [{"membro_id":1,"inicio":"08:00","fim":"12:00"},...]
	Status             string  `json:"status"`     // pendente|confirmado|rejeitado|cancelado
	Note               string  `json:"observacao"`
	CreatedAt          string  `json:"criado_em"`
	ConfirmedAt        *string `json:"confirmado_em"`
}

const excCols = "id, date, type, original_member_id, substitute_member_id, swap_date, end_date, start_time, end_time, slots_json, status, note, created_at, confirmed_at"

func scanException(row interface{ Scan(...any) error }) (Exception, error) {
	var e Exception
	err := row.Scan(&e.ID, &e.Date, &e.Type, &e.OriginalMemberID, &e.SubstituteMemberID,
		&e.SwapDate, &e.EndDate, &e.StartTime, &e.EndTime, &e.SlotsJSON, &e.Status, &e.Note, &e.CreatedAt, &e.ConfirmedAt)
	return e, err
}

// ListExceptionsForRange retorna exceções cuja date OU swap_date cai em
// [from, to], ou cujo período [date, end_date] intersecta o intervalo.
func (s *Store) ListExceptionsForRange(from, to string, statuses []string) ([]Exception, error) {
	q := "SELECT " + excCols + " FROM exceptions WHERE ((date BETWEEN ? AND ?) OR (swap_date IS NOT NULL AND swap_date BETWEEN ? AND ?) OR (end_date IS NOT NULL AND date <= ? AND end_date >= ?))"
	args := []any{from, to, from, to, to, from}
	if len(statuses) > 0 {
		q += " AND status IN (?" + repeat(",?", len(statuses)-1) + ")"
		for _, st := range statuses {
			args = append(args, st)
		}
	}
	q += " ORDER BY date, start_time"
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Exception{}
	for rows.Next() {
		e, err := scanException(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func (s *Store) GetException(id int64) (Exception, error) {
	row := s.DB.QueryRow("SELECT "+excCols+" FROM exceptions WHERE id = ?", id)
	return scanException(row)
}

func (s *Store) CreateException(e Exception) (Exception, error) {
	res, err := s.DB.Exec(
		"INSERT INTO exceptions (date, type, original_member_id, substitute_member_id, swap_date, end_date, start_time, end_time, slots_json, status, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pendente', ?)",
		e.Date, e.Type, e.OriginalMemberID, e.SubstituteMemberID, e.SwapDate, e.EndDate, e.StartTime, e.EndTime, e.SlotsJSON, e.Note)
	if err != nil {
		return Exception{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetException(id)
}

func (s *Store) SetExceptionStatus(id int64, status string) (Exception, error) {
	if status == "confirmado" {
		_, err := s.DB.Exec("UPDATE exceptions SET status = ?, confirmed_at = datetime('now') WHERE id = ?", status, id)
		if err != nil {
			return Exception{}, err
		}
	} else {
		_, err := s.DB.Exec("UPDATE exceptions SET status = ? WHERE id = ?", status, id)
		if err != nil {
			return Exception{}, err
		}
	}
	return s.GetException(id)
}
