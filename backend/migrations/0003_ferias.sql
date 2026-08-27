-- +goose Up
-- Tipo 'ferias': período [date, end_date] em que os plantões do integrante
-- são mascarados (com substituto opcional assumindo-os). O rodízio continua
-- correndo por baixo, então no retorno tudo volta exatamente como era.
CREATE TABLE exceptions_new (
  id INTEGER PRIMARY KEY,
  date TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('substituicao','troca','extra','ausencia','feriado','edicao_dia','ferias')),
  original_member_id INTEGER REFERENCES members(id),
  substitute_member_id INTEGER REFERENCES members(id),
  swap_date TEXT,
  end_date TEXT,
  start_time TEXT NOT NULL DEFAULT '00:00',
  end_time TEXT NOT NULL DEFAULT '23:59',
  slots_json TEXT,
  status TEXT NOT NULL DEFAULT 'pendente'
    CHECK (status IN ('pendente','confirmado','rejeitado','cancelado')),
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  confirmed_at TEXT
);
INSERT INTO exceptions_new (id, date, type, original_member_id, substitute_member_id,
  swap_date, start_time, end_time, slots_json, status, note, created_at, confirmed_at)
  SELECT id, date, type, original_member_id, substitute_member_id,
    swap_date, start_time, end_time, slots_json, status, note, created_at, confirmed_at
  FROM exceptions;
DROP TABLE exceptions;
ALTER TABLE exceptions_new RENAME TO exceptions;
CREATE INDEX idx_exceptions_date ON exceptions(date);
CREATE INDEX idx_exceptions_swap_date ON exceptions(swap_date);
CREATE INDEX idx_exceptions_end_date ON exceptions(end_date);

-- +goose Down
-- sem down: recriar o CHECK antigo perderia o tipo 'ferias'
