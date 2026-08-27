// Package api expõe a API REST e serve o frontend embutido.
package api

import (
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/config"
	"github.com/joao/plantoes/internal/notify"
	"github.com/joao/plantoes/internal/store"
	"github.com/joao/plantoes/internal/web"
)

type Server struct {
	St    *store.Store
	Cfg   *config.Config
	Tg    *notify.Telegram
	Sched *notify.Scheduler
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api := r.Group("/api")
	{
		// Rotas públicas: só a visualização do calendário e o login
		api.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
		api.GET("/calendario/:ano/:mes", s.getCalendar)
		api.POST("/login", s.login)
	}

	priv := api.Group("", s.authRequired())
	{
		priv.POST("/logout", s.logout)
		priv.GET("/me", s.me)
		priv.POST("/senha", s.changePassword)

		priv.GET("/membros", s.listMembers)
		priv.POST("/membros", s.createMember)
		priv.GET("/membros/:id", s.getMember)
		priv.PUT("/membros/:id", s.updateMember)
		priv.DELETE("/membros/:id", s.deactivateMember)

		priv.GET("/membros/:id/regras", s.listRules)
		priv.PUT("/membros/:id/regras", s.replaceRules)
		priv.DELETE("/regras/:id", s.deleteRule)

		priv.POST("/excecoes", s.createException)
		priv.GET("/excecoes", s.listExceptions)
		priv.POST("/excecoes/:id/confirmar", s.confirmException)
		priv.POST("/excecoes/:id/rejeitar", s.rejectException)
		priv.DELETE("/excecoes/:id", s.cancelException)

		priv.GET("/historico", s.listAudit)
		priv.POST("/alertas/teste", s.testAlert)
	}

	web.Register(r)
	return r
}

var (
	timeRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

func validTime(t string) bool { return timeRe.MatchString(t) }
func validDate(d string) bool {
	if !dateRe.MatchString(d) {
		return false
	}
	_, err := time.Parse("2006-01-02", d)
	return err == nil
}

func (s *Server) today() string {
	return time.Now().In(s.Cfg.Location).Format("2006-01-02")
}

func badRequest(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"erro": msg})
}

func serverError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"erro": err.Error()})
}
