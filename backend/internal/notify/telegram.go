// Package notify envia alertas de plantão para um grupo do Telegram.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Telegram struct {
	BotToken string
	ChatID   string
	client   *http.Client
}

func NewTelegram(botToken, chatID string) *Telegram {
	return &Telegram{
		BotToken: botToken,
		ChatID:   chatID,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *Telegram) Enabled() bool {
	return t.BotToken != "" && t.ChatID != ""
}

// Send faz o equivalente a:
//
//	curl -X POST https://api.telegram.org/bot<token>/sendMessage \
//	     -d chat_id=<chat_id> -d text=<mensagem>
//
// com uma nova tentativa em caso de falha.
func (t *Telegram) Send(text string) error {
	if !t.Enabled() {
		log.Printf("[telegram] desabilitado (sem token/chat_id); mensagem: %s", text)
		return nil
	}
	body, _ := json.Marshal(map[string]string{"chat_id": t.ChatID, "text": text})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)

	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		resp, err := t.client.Post(url, "application/json", bytes.NewReader(body))
		if err != nil {
			lastErr = err
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return nil
		}
		lastErr = fmt.Errorf("telegram respondeu HTTP %d", resp.StatusCode)
		time.Sleep(2 * time.Second)
	}
	return lastErr
}
