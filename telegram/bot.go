package bot

import (
	"LapaTelegramBot/schedule"
	"LapaTelegramBot/zabbix"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API             *tgbotapi.BotAPI
	Zabbix          *zabbix.Client
	ScheduleStore   *schedule.Storage
	ScheduleManager *schedule.Manager
	Commands        map[string]func(tgbotapi.Update)
	AllowedChats    map[int64]bool
}

func StartBot() {
	token := os.Getenv("TELEGRAM_API_TOKEN")
	chatsIds := os.Getenv("TELEGRAM_ALLOWED_CHAT_ID")

	allowedChatID := strings.Split(chatsIds, ",")
	allowed := loadAllowedChats(allowedChatID)

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	bot := &Bot{
		API:          api,
		Zabbix:       zabbix.NewClient(),
		AllowedChats: allowed,
	}

	bot.initCommands()
	bot.initSchedule()

	log.Println("Bot iniciado como:", bot.API.Self.UserName)
	bot.Start()
}

func (b *Bot) initSchedule() {
	b.ScheduleStore = schedule.NewStorage()
	b.ScheduleManager = schedule.NewManager()

	b.ScheduleStore.Load()

	// Passa o método ExecuteCommand como callback
	schedule.LoadExistingJobs(b.ScheduleStore, b.ScheduleManager, b.ExecuteCommand)

	b.ScheduleManager.Start()
}

func (b *Bot) initCommands() {
	b.Commands = map[string]func(tgbotapi.Update){
		"status_check":     b.handleStatusCheck,
		"protheus_status":  b.handleProtheusStatus,
		"listip":           b.handleListIp,
		"ping":             b.handlePing,
		"services":         b.handleRemoteServices,
		"printers_counter": b.handlePrinterCounter,
		"schedule_add":     b.handleScheduleAdd,
		"schedule_remove":  b.handleScheduleRemove,
		"schedule_list":    b.handleScheduleList,
		"schedule_help":    b.handleScheduleHelp,
		"restart_win":      b.handleRestartWindowsHost,
		"shutdown_win":     b.handleShutdownWindowsHost,
	}
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		logUpdate(update)
		// Ignora qualquer update sem mensagem ou mensagens sem ChatID autorizados
		if update.Message == nil {
			continue
		}

		if !b.AllowedChats[update.Message.Chat.ID] {
			continue
		}

		if !update.Message.IsCommand() { // Ignora mensagens que não são comandos
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe um comando.")
			b.API.Send(msg)
			continue
		}

		cmd := update.Message.Command()
		if handler, ok := b.Commands[cmd]; ok {
			handler(update)
		}
	}
}

func (b *Bot) ExecuteCommand(cmd string, chatID int64) {
	// Remove a barra inicial se existir (embora o scheduler geralmente guarde o comando raw)
	cmdClean := strings.TrimPrefix(cmd, "/")
	parts := strings.Split(cmdClean, " ")
	commandName := parts[0]

	// Cria um Update fake para reaproveitar a lógica dos handlers
	fakeUpdate := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: chatID},
			Text: "/" + cmdClean, // Recria o texto do comando
			Entities: []tgbotapi.MessageEntity{
				{Type: "bot_command", Offset: 0, Length: len(commandName) + 1},
			},
		},
	}

	if handler, ok := b.Commands[commandName]; ok {
		log.Printf("Executando handler via scheduler para comando: %s", commandName)
		handler(fakeUpdate)
	} else {
		log.Printf("Comando não encontrado: %s", commandName)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Comando '%s' não encontrado", commandName))
		b.API.Send(msg)
	}
}

func logUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}
	name := update.Message.From.FirstName
	if update.Message.From.LastName != "" {
		name += " " + update.Message.From.LastName
	}

	username := update.Message.Chat.UserName
	if username == "" {
		username = "[N/A]"
	}

	comando := update.Message.Text

	log.Printf(
		"\n• Chat ID:%d\n• Nome:%s\n• Username:%s\n• Comando:%s\n\n",
		update.Message.Chat.ID,
		name,
		username,
		comando,
	)
}

func loadAllowedChats(parts []string) map[int64]bool {
	allowed := make(map[int64]bool)

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Printf("⚠️  Ignorando ID inválido no ALLOWED_CHATS: %s\n", p)
			continue
		}

		allowed[id] = true
	}

	return allowed
}
