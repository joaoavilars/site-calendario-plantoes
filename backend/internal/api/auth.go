package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/store"
)

const sessionTTLHours = 7 * 24

func bearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

// authRequired protege as rotas administrativas: exige um token de sessão
// válido no header Authorization.
func (s *Server) authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := bearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "não autenticado"})
			return
		}
		id, username, ok := s.St.SessionAdmin(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"erro": "sessão inválida ou expirada — faça login novamente"})
			return
		}
		c.Set("adminID", id)
		c.Set("adminUser", username)
		c.Set("adminToken", token)
		c.Next()
	}
}

func (s *Server) login(c *gin.Context) {
	var in struct {
		Usuario string `json:"usuario"`
		Senha   string `json:"senha"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "corpo inválido")
		return
	}
	id, err := s.St.Authenticate(strings.TrimSpace(in.Usuario), in.Senha)
	if errors.Is(err, store.ErrCredenciais) {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": "usuário ou senha inválidos"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}
	token, err := s.St.CreateSession(id, sessionTTLHours)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "usuario": strings.TrimSpace(in.Usuario)})
}

func (s *Server) logout(c *gin.Context) {
	s.St.DeleteSession(c.GetString("adminToken"))
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *Server) me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"usuario": c.GetString("adminUser")})
}

func (s *Server) changePassword(c *gin.Context) {
	var in struct {
		SenhaAtual string `json:"senha_atual"`
		SenhaNova  string `json:"senha_nova"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "corpo inválido")
		return
	}
	if len(in.SenhaNova) < 6 {
		badRequest(c, "a nova senha deve ter pelo menos 6 caracteres")
		return
	}
	adminID := c.GetInt64("adminID")
	err := s.St.ChangePassword(adminID, in.SenhaAtual, in.SenhaNova, c.GetString("adminToken"))
	if errors.Is(err, store.ErrCredenciais) {
		badRequest(c, "senha atual incorreta")
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}
	s.St.Audit("admin", adminID, "password_change", gin.H{"usuario": c.GetString("adminUser")})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
