package bot

import (
	"fmt"
	"log"
	"strings"
	"time"

	"LapaTelegramBot/monitor"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Monitor struct {
	ChatID          int64
	IntervalMinutes int
	stopCh          chan struct{}
	updateInterval  chan int
	waitingInterval bool
	lastMsgID       int
}

func NewMonitor(chatID int64, minutes int) *Monitor {
	return &Monitor{ChatID: chatID, IntervalMinutes: minutes, stopCh: make(chan struct{}, 1), updateInterval: make(chan int, 1)}
}

func (m *Monitor) run(b *Bot) {
	exclude := []string{"Applications", "Impressoras"}

	// Função que realiza a checagem e envia notificação se necessário
	doCheck := func() {
		hosts, err := monitor.CheckHostsStatusExcludingGroups(b.Zabbix, exclude)
		if err != nil {
			log.Printf("Erro ao checar hosts no monitor: %v", err)
			return
		}

		var offline []string
		for _, h := range hosts {
			if strings.HasPrefix(h, "❌") {
				offline = append(offline, h)
			}
		}

		if len(offline) == 0 {
			// Nenhuma ação se todos online
			return
		}

		// Envia mensagem com botões; antes apaga a última mensagem enviada pelo monitor
		if m.lastMsgID != 0 {
			del := tgbotapi.NewDeleteMessage(m.ChatID, m.lastMsgID)
			b.API.Send(del)
			m.lastMsgID = 0
		}

		msgText := "❗ Hosts offline encontrados:\n\n"
		for _, o := range offline {
			msgText += o + "\n"
		}

		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⏱️ Aumentar intervalo", fmt.Sprintf("monitor:increase:%d", m.ChatID)),
				tgbotapi.NewInlineKeyboardButtonData("🛑 Parar aviso", fmt.Sprintf("monitor:stop:%d", m.ChatID)),
			),
		)

		msg := tgbotapi.NewMessage(m.ChatID, msgText)
		msg.ReplyMarkup = kb
		sent, err := b.API.Send(msg)
		if err == nil {
			m.lastMsgID = sent.MessageID
		}
	}

	// Execute a primeira checagem imediatamente
	doCheck()

	ticker := time.NewTicker(time.Duration(m.IntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			log.Printf("Monitor stopped for chat %d", m.ChatID)
			return
		case newInterval := <-m.updateInterval:
			if newInterval <= 0 {
				continue
			}
			m.IntervalMinutes = newInterval
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(m.IntervalMinutes) * time.Minute)
			log.Printf("Monitor interval updated to %d minutes for chat %d", m.IntervalMinutes, m.ChatID)
		case <-ticker.C:
			doCheck()
		}
	}
}
