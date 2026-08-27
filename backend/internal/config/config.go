package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Timezone string `yaml:"timezone"`
	HTTP     struct {
		Porta int `yaml:"porta"`
	} `yaml:"http"`
	Banco struct {
		Caminho string `yaml:"caminho"`
	} `yaml:"banco"`
	Telegram struct {
		BotToken string `yaml:"bot_token"`
		ChatID   string `yaml:"chat_id"`
	} `yaml:"telegram"`
	Alertas struct {
		ResumoDiario struct {
			Habilitado bool   `yaml:"habilitado"`
			Horario    string `yaml:"horario"`
		} `yaml:"resumo_diario"`
		LembretePlantao struct {
			Habilitado          bool `yaml:"habilitado"`
			AntecedenciaMinutos int  `yaml:"antecedencia_minutos"`
		} `yaml:"lembrete_plantao"`
		AvisoTrocaConfirmada struct {
			Habilitado bool `yaml:"habilitado"`
		} `yaml:"aviso_troca_confirmada"`
	} `yaml:"alertas"`

	Location *time.Location `yaml:"-"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	cfg.Timezone = "America/Sao_Paulo"
	cfg.HTTP.Porta = 8080
	cfg.Banco.Caminho = "./data/plantoes.db"
	cfg.Alertas.ResumoDiario.Horario = "07:30"
	cfg.Alertas.LembretePlantao.AntecedenciaMinutos = 15

	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	if v := os.Getenv("TELEGRAM_BOT_TOKEN"); v != "" {
		cfg.Telegram.BotToken = v
	}
	if v := os.Getenv("TELEGRAM_CHAT_ID"); v != "" {
		cfg.Telegram.ChatID = v
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.Banco.Caminho = v
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return nil, fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
	}
	cfg.Location = loc
	return cfg, nil
}
