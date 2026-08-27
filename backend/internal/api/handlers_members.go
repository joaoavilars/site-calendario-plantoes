package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/store"
)

func paramID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		badRequest(c, "id inválido")
		return 0, false
	}
	return id, true
}

func (s *Server) listMembers(c *gin.Context) {
	includeInactive := c.Query("todos") == "1"
	members, err := s.St.ListMembers(includeInactive)
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}

func (s *Server) getMember(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	m, err := s.St.GetMember(id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"erro": "integrante não encontrado"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}
	c.JSON(http.StatusOK, m)
}

type memberInput struct {
	Nome            string `json:"nome"`
	Cor             string `json:"cor"`
	MencaoTelegram  string `json:"mencao_telegram"`
	Ativo           *bool  `json:"ativo"`
}

func (s *Server) createMember(c *gin.Context) {
	var in memberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "corpo inválido: "+err.Error())
		return
	}
	in.Nome = strings.TrimSpace(in.Nome)
	if in.Nome == "" {
		badRequest(c, "nome é obrigatório")
		return
	}
	if in.Cor == "" {
		in.Cor = "#3b82f6"
	}
	m, err := s.St.CreateMember(store.Member{Name: in.Nome, Color: in.Cor, TelegramMention: in.MencaoTelegram})
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			badRequest(c, "já existe um integrante com esse nome")
			return
		}
		serverError(c, err)
		return
	}
	s.St.Audit("member", m.ID, "create", m)
	c.JSON(http.StatusCreated, m)
}

func (s *Server) updateMember(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	current, err := s.St.GetMember(id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"erro": "integrante não encontrado"})
		return
	}
	if err != nil {
		serverError(c, err)
		return
	}
	var in memberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		badRequest(c, "corpo inválido: "+err.Error())
		return
	}
	if strings.TrimSpace(in.Nome) != "" {
		current.Name = strings.TrimSpace(in.Nome)
	}
	if in.Cor != "" {
		current.Color = in.Cor
	}
	current.TelegramMention = in.MencaoTelegram
	if in.Ativo != nil {
		current.Active = *in.Ativo
	}
	m, err := s.St.UpdateMember(current)
	if err != nil {
		serverError(c, err)
		return
	}
	s.St.Audit("member", m.ID, "update", m)
	c.JSON(http.StatusOK, m)
}

func (s *Server) deactivateMember(c *gin.Context) {
	id, ok := paramID(c)
	if !ok {
		return
	}
	if err := s.St.DeactivateMember(id, s.today()); err != nil {
		serverError(c, err)
		return
	}
	s.St.Audit("member", id, "deactivate", gin.H{"id": id})
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
