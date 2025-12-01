package bot

import (
	"TelegramNotify/monitor"
	"TelegramNotify/zabbix"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type CommandHandler func(bot *tgbotapi.BotAPI, update tgbotapi.Update)

var commands = map[string]CommandHandler{
	"ping":             handlePing,
	"status_check":     handleStatusCheck,
	"printers_counter": handlePrinterCounter,
	"restart_win":      handleRestartWindowsHost,
	"shutdown_win":     handleShutdownWindowsHost,
	"listip":           handleListIp,
	"protheus_status":  handleProtheusStatus,
}

func StartBot() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	var env map[string]string
	env, e := godotenv.Read()

	if e != nil {
		log.Fatal(e)
	}

	token := env["TELEGRAM_API_TOKEN"]
	chatsIds := env["TELEGRAM_ALLOWED_CHAT_ID"]

	allowedChatID := strings.Split(chatsIds, ",")

	allowed := loadAllowedChats(allowedChatID)

	if err != nil {
		log.Fatal("Não foi possível carregar os IDs dos chats permitidos.", err)
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	log.Println("Bot iniciado como:", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		logar(update)
		// Ignora qualquer update sem mensagem ou mensagens sem ChatID autorizados
		if update.Message == nil {
			continue
		}

		if !allowed[update.Message.Chat.ID] {
			continue
		}

		if !update.Message.IsCommand() { // Ignora mensagens que não são comandos
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe um comando.")
			bot.Send(msg)
			continue
		}

		cmd := update.Message.Command()
		if handler, ok := commands[cmd]; ok {
			handler(bot, update)
		}
	}
}

func logar(update tgbotapi.Update) {
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
		"\n• Chat ID:%d\n• Nome:%s\n• Username:%s\n• Comando:%s\n",
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

func handlePing(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o IP. Ex: /ping 192.168.0.1")
		bot.Send(msg)
		return
	}

	var wg sync.WaitGroup
	result := make(chan string)

	chatID := update.Message.Chat.ID
	for i := 1; i < len(parts); i++ {
		ip := parts[i]

		wg.Add(1)
		go pingFunc(ip, &wg, result)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	for resultText := range result {
		msg := tgbotapi.NewMessage(chatID, resultText)
		bot.Send(msg)
	}
}

func pingFunc(ip string, wg *sync.WaitGroup, channel chan string) {
	defer wg.Done()
	pinger, err := ping.NewPinger(ip)
	if err != nil {
		channel <- err.Error()
		return
	}

	pinger.Count = 3
	pinger.Interval = 300 * time.Millisecond
	pinger.Timeout = 3 * time.Second

	if runtime.GOOS == "windows" {
		pinger.SetPrivileged(true) /* Falha no Windows caso o programa não seja executado como administrador */
	}

	err = pinger.Run()
	if err != nil {
		// Erro típico de host offline no Windows
		if strings.Contains(strings.ToLower(err.Error()), "wsarecvfrom") {
			channel <- fmt.Sprintf("❌ %s\nStatus: OFFLINE (nenhuma resposta)", ip)
		} else {
			channel <- fmt.Sprintf("❌ %s\nErro no ping: %v", ip, err)
		}
		return
	}

	stats := pinger.Statistics()

	response := fmt.Sprintf(
		"✅ %s\nEnviados: %d | Recebidos: %d | Perda: %.0f%%\nLatência média: %v",
		ip,
		stats.PacketsSent,
		stats.PacketsRecv,
		stats.PacketLoss,
		stats.AvgRtt,
	)
	channel <- response
}

func handleStatusCheck(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	z := zabbix.NewClient()
	hosts, err := monitor.CheckHostsStatus(z)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar Zabbix:\n%v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		log.Println(err)
		return
	}

	msg := "🚥🚥🚥🚥🚥🚥 Status dos Hosts 🚥🚥🚥🚥🚥🚥\n\n"
	for _, h := range hosts {
		msg += h + "\n"
	}

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func handlePrinterCounter(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	z := zabbix.NewClient()
	printers, err := monitor.GetPrintersCounter(z)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar Zabbix:\n%v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		log.Println(err)
		return
	}

	msg := "🔢🔢🔢🔢🔢🔢 CONTADORES 🔢🔢🔢🔢🔢🔢\n\n"
	for _, printer := range printers {
		msg += "====== " + printer.HostData.Host + " ======\n"
		if printer.BlackCounter != 0 {
			msg += fmt.Sprintf("Preto e Branco: %d\n", printer.BlackCounter)
		}
		if printer.ColorCounter != 0 {
			msg += fmt.Sprintf("Colorido: %d\n", printer.ColorCounter)
		}
		if printer.TotalCounter != 0 {
			msg += fmt.Sprintf("Total: %d\n", printer.TotalCounter)
		}
		msg += "\n"
	}

	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func handleRestartWindowsHost(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o hostname. Ex: /restart_win \\\\LVMAQUINA")
		bot.Send(msg)
		return
	}

	host := parts[1]
	log.Println("Handler restart_win acionado, destino: %s", host)

	cmd := exec.Command(
		"shutdown",
		"/r",
		"/t", "0",
		"/m", fmt.Sprintf("\\\\%s", host),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		e := fmt.Sprintf("Erro ao tentar reiniciar %s: %v\nSaída: %s", host, err, string(output))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, e)
		bot.Send(msg)
		return
	}
	m := fmt.Sprintf("✅ Comando executado para: %s", host)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
	bot.Send(msg)
}

func handleShutdownWindowsHost(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o hostname. Ex: /restart_w_host \\\\LVMAQUINA")
		bot.Send(msg)
		return
	}
	host := parts[1]
	log.Println("Handler shutdown_w_host acionado, destino: %s", host)

	cmd := exec.Command(
		"shutdown",
		"/s",
		"/t", "0",
		"/m", fmt.Sprintf("\\\\%s", host),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		e := fmt.Sprintf("Erro ao tentar desligar %s: %v\nSaída: %s", host, err, string(output))
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, e)
		bot.Send(msg)
		return
	}
	m := fmt.Sprintf("✅ Comando executado para: %s", host)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
	bot.Send(msg)
}

func handleListIp(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	z := zabbix.NewClient()
	hostsList, err := z.ListIps()
	if err != nil {
		msg := fmt.Sprintf("Erro ao listar Zabbix:\n%v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		return
	}

	msg := "🌐🌐🌐 Lista de IPs 🌐🌐🌐\n\n"

	for _, host := range hostsList {
		msg += fmt.Sprintf("🌐\t%s\n", host.Host)
		for i := range host.Interfaces {
			msg += fmt.Sprintf("•\t%s\n", host.Interfaces[i].IP)
		}
		msg += "\n"
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func handleProtheusStatus(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	z := zabbix.NewClient()
	services, err := z.GetProtheusServiceStatus()
	if err != nil {
		msg := fmt.Sprintf("Erro ao pegar os status:\n%v", err)
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		return
	}

	msg := "⚙️⚙️⚙️ Protheus Services Status ⚙️⚙️⚙️\n\n"
	for _, service := range services {
		name := ""
		re := regexp.MustCompile(`"([^"]+)"`)
		match := re.FindStringSubmatch(service.Name)

		if len(match) > 1 {
			name = match[1]
		} else {
			name = service.Name
		}

		if service.Lastvalue == "0" && service.Prevvalue == "0" {
			msg += fmt.Sprintf("✅ %s\n", name)
		} else {
			msg += fmt.Sprintf("❌ %s\n", name)
		}
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}
