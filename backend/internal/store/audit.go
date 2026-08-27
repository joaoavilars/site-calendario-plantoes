package store

import "encoding/json"

type AuditEntry struct {
	ID        int64           `json:"id"`
	Entity    string          `json:"entidade"`
	EntityID  int64           `json:"entidade_id"`
	Action    string          `json:"acao"`
	Payload   json.RawMessage `json:"dados"`
	CreatedAt string          `json:"criado_em"`
}

func (s *Store) Audit(entity string, entityID int64, action string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		data = []byte("{}")
	}
	// Auditoria não deve derrubar a operação principal; erro é ignorado de propósito.
	s.DB.Exec("INSERT INTO audit_log (entity, entity_id, action, payload_json) VALUES (?, ?, ?, ?)",
		entity, entityID, action, string(data))
}

func (s *Store) ListAudit(entity, from, to string, limit int) ([]AuditEntry, error) {
	q := "SELECT id, entity, entity_id, action, payload_json, created_at FROM audit_log WHERE 1=1"
	args := []any{}
	if entity != "" {
		q += " AND entity = ?"
		args = append(args, entity)
	}
	if from != "" {
		q += " AND created_at >= ?"
		args = append(args, from)
	}
	if to != "" {
		q += " AND created_at <= ? || ' 23:59:59'"
		args = append(args, to)
	}
	q += " ORDER BY id DESC LIMIT ?"
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	args = append(args, limit)
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AuditEntry{}
	for rows.Next() {
		var e AuditEntry
		var payload string
		if err := rows.Scan(&e.ID, &e.Entity, &e.EntityID, &e.Action, &payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = json.RawMessage(payload)
		out = append(out, e)
	}
	return out, rows.Err()
}

// MarkAlertSent registra a chave do alerta; retorna false se já havia sido enviado.
func (s *Store) MarkAlertSent(key string) (bool, error) {
	res, err := s.DB.Exec("INSERT OR IGNORE INTO alerts_sent (alert_key) VALUES (?)", key)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
