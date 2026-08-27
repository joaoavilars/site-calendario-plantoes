package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/joao/plantoes/internal/api"
	"github.com/joao/plantoes/internal/config"
	"github.com/joao/plantoes/internal/notify"
	"github.com/joao/plantoes/internal/store"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	st, err := store.Open(cfg.Banco.Caminho)
	if err != nil {
		log.Fatalf("banco: %v", err)
	}
	defer st.Close()

	if created, err := st.EnsureDefaultAdmin(); err != nil {
		log.Fatalf("admin inicial: %v", err)
	} else if created {
		log.Printf("⚠️  Usuário administrador criado: admin / admin123 — troque a senha na engrenagem ⚙️ após o primeiro login")
	}

	tg := notify.NewTelegram(cfg.Telegram.BotToken, cfg.Telegram.ChatID)
	sched := notify.NewScheduler(cfg, st, tg)
	if err := sched.Start(); err != nil {
		log.Fatalf("scheduler de alertas: %v", err)
	}
	defer sched.Stop()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}
	server := &api.Server{St: st, Cfg: cfg, Tg: tg, Sched: sched}
	addr := fmt.Sprintf(":%d", cfg.HTTP.Porta)
	log.Printf("Calendário de Plantões ouvindo em %s (tz %s)", addr, cfg.Timezone)
	if err := server.Router().Run(addr); err != nil {
		log.Fatalf("http: %v", err)
	}
}
