package store

import "database/sql"

type Member struct {
	ID              int64  `json:"id"`
	Name            string `json:"nome"`
	Color           string `json:"cor"`
	TelegramMention string `json:"mencao_telegram"`
	Active          bool   `json:"ativo"`
	CreatedAt       string `json:"criado_em"`
}

func scanMember(row interface{ Scan(...any) error }) (Member, error) {
	var m Member
	var active int
	err := row.Scan(&m.ID, &m.Name, &m.Color, &m.TelegramMention, &active, &m.CreatedAt)
	m.Active = active == 1
	return m, err
}

func (s *Store) ListMembers(includeInactive bool) ([]Member, error) {
	q := "SELECT id, name, color, telegram_mention, active, created_at FROM members"
	if !includeInactive {
		q += " WHERE active = 1"
	}
	q += " ORDER BY name"
	rows, err := s.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Member{}
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) GetMember(id int64) (Member, error) {
	row := s.DB.QueryRow("SELECT id, name, color, telegram_mention, active, created_at FROM members WHERE id = ?", id)
	return scanMember(row)
}

func (s *Store) CreateMember(m Member) (Member, error) {
	res, err := s.DB.Exec("INSERT INTO members (name, color, telegram_mention) VALUES (?, ?, ?)",
		m.Name, m.Color, m.TelegramMention)
	if err != nil {
		return Member{}, err
	}
	id, _ := res.LastInsertId()
	return s.GetMember(id)
}

func (s *Store) UpdateMember(m Member) (Member, error) {
	active := 0
	if m.Active {
		active = 1
	}
	_, err := s.DB.Exec("UPDATE members SET name = ?, color = ?, telegram_mention = ?, active = ? WHERE id = ?",
		m.Name, m.Color, m.TelegramMention, active, m.ID)
	if err != nil {
		return Member{}, err
	}
	return s.GetMember(m.ID)
}

// DeactivateMember é o "delete": desativa e encerra as regras vigentes.
func (s *Store) DeactivateMember(id int64, today string) error {
	if _, err := s.DB.Exec("UPDATE members SET active = 0 WHERE id = ?", id); err != nil {
		return err
	}
	_, err := s.DB.Exec(
		"UPDATE rotation_rules SET effective_to = ? WHERE member_id = ? AND (effective_to IS NULL OR effective_to > ?)",
		today, id, today)
	return err
}

var ErrNotFound = sql.ErrNoRows
