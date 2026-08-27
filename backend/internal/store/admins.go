package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

var ErrCredenciais = errors.New("usuário ou senha inválidos")

// EnsureDefaultAdmin cria o usuário 'admin' com senha 'admin123' se ainda
// não existir nenhum administrador. Retorna true se criou (para logar o aviso).
func (s *Store) EnsureDefaultAdmin() (bool, error) {
	var n int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM admins").Scan(&n); err != nil {
		return false, err
	}
	if n > 0 {
		return false, nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	_, err = s.DB.Exec("INSERT INTO admins (username, password_hash) VALUES ('admin', ?)", string(hash))
	return err == nil, err
}

// Authenticate valida usuário/senha e retorna o id do admin.
func (s *Store) Authenticate(username, password string) (int64, error) {
	var id int64
	var hash string
	err := s.DB.QueryRow("SELECT id, password_hash FROM admins WHERE username = ?", username).Scan(&id, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrCredenciais
	}
	if err != nil {
		return 0, err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return 0, ErrCredenciais
	}
	return id, nil
}

// CreateSession gera um token de sessão com validade em horas.
func (s *Store) CreateSession(adminID int64, ttlHours int) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	// limpeza oportunista de sessões vencidas
	s.DB.Exec("DELETE FROM sessions WHERE expires_at <= datetime('now')")
	_, err := s.DB.Exec(
		"INSERT INTO sessions (token, admin_id, expires_at) VALUES (?, ?, datetime('now', ?))",
		token, adminID, "+"+strconv.Itoa(ttlHours)+" hours")
	if err != nil {
		return "", err
	}
	return token, nil
}

// SessionAdmin resolve um token válido para (adminID, username).
func (s *Store) SessionAdmin(token string) (int64, string, bool) {
	var id int64
	var username string
	err := s.DB.QueryRow(
		"SELECT a.id, a.username FROM sessions s JOIN admins a ON a.id = s.admin_id WHERE s.token = ? AND s.expires_at > datetime('now')",
		token).Scan(&id, &username)
	return id, username, err == nil
}

func (s *Store) DeleteSession(token string) {
	s.DB.Exec("DELETE FROM sessions WHERE token = ?", token)
}

// ChangePassword troca a senha do admin após validar a atual.
// Invalida as demais sessões (mantém a atual, se informada).
func (s *Store) ChangePassword(adminID int64, current, newPassword, keepToken string) error {
	var hash string
	if err := s.DB.QueryRow("SELECT password_hash FROM admins WHERE id = ?", adminID).Scan(&hash); err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(current)) != nil {
		return ErrCredenciais
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := s.DB.Exec("UPDATE admins SET password_hash = ? WHERE id = ?", string(newHash), adminID); err != nil {
		return err
	}
	s.DB.Exec("DELETE FROM sessions WHERE admin_id = ? AND token != ?", adminID, keepToken)
	return nil
}
