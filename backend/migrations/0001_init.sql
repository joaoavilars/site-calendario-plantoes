-- +goose Up
CREATE TABLE members (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  color TEXT NOT NULL DEFAULT '#3b82f6',
  telegram_mention TEXT NOT NULL DEFAULT '',
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE rotation_rules (
  id INTEGER PRIMARY KEY,
  member_id INTEGER NOT NULL REFERENCES members(id),
  weekday INTEGER NOT NULL CHECK (weekday BETWEEN 0 AND 6),
  interval_weeks INTEGER NOT NULL DEFAULT 1 CHECK (interval_weeks >= 1),
  start_time TEXT NOT NULL,
  end_time TEXT NOT NULL,
  effective_from TEXT NOT NULL,
  effective_to TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_rules_member ON rotation_rules(member_id);

CREATE TABLE exceptions (
  id INTEGER PRIMARY KEY,
  date TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('substituicao','troca','extra','ausencia')),
  original_member_id INTEGER REFERENCES members(id),
  substitute_member_id INTEGER REFERENCES members(id),
  swap_date TEXT,
  start_time TEXT NOT NULL,
  end_time TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pendente'
    CHECK (status IN ('pendente','confirmado','rejeitado','cancelado')),
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  confirmed_at TEXT
);
CREATE INDEX idx_exceptions_date ON exceptions(date);
CREATE INDEX idx_exceptions_swap_date ON exceptions(swap_date);

CREATE TABLE audit_log (
  id INTEGER PRIMARY KEY,
  entity TEXT NOT NULL,
  entity_id INTEGER NOT NULL,
  action TEXT NOT NULL,
  payload_json TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_audit_entity ON audit_log(entity, entity_id);

CREATE TABLE alerts_sent (
  id INTEGER PRIMARY KEY,
  alert_key TEXT NOT NULL UNIQUE,
  sent_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- +goose Down
DROP TABLE alerts_sent;
DROP TABLE audit_log;
DROP TABLE exceptions;
DROP TABLE rotation_rules;
DROP TABLE members;
