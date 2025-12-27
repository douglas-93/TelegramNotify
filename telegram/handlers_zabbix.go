package bot

import (
	"LapaTelegramBot/file_handler"
	"LapaTelegramBot/monitor"
	"fmt"
	"log"
	"os"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStatusCheck(update tgbotapi.Update) {
	hosts, err := monitor.CheckHostsStatus(b.Zabbix)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar Zabbix:\n%v", err)
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		log.Println(err)
		return
	}

	msg := "🚥🚥🚥 Status dos Hosts 🚥🚥🚥\n\n"
	for _, h := range hosts {
		msg += h + "\n"
	}

	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func (b *Bot) handlePrinterCounter(update tgbotapi.Update) {
	printers, err := monitor.GetPrintersCounter(b.Zabbix)
	if err != nil {
		msg := fmt.Sprintf("Erro ao consultar Zabbix:\n%v", err)
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		log.Println(err)
		return
	}

	msg := "🔢🔢🔢 CONTADORES 🔢🔢🔢\n\n"
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

	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
	excelFile := file_handler.GenerateSheet()
	b.API.Send(tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath(excelFile)))
	os.Remove(excelFile)
}

func (b *Bot) handleListIp(update tgbotapi.Update) {
	hostsList, err := b.Zabbix.ListIps()
	if err != nil {
		msg := fmt.Sprintf("Erro ao listar Zabbix:\n%v", err)
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
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
	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}

func (b *Bot) handleProtheusStatus(update tgbotapi.Update) {
	services, err := b.Zabbix.GetProtheusServiceStatus()
	if err != nil {
		msg := fmt.Sprintf("Erro ao pegar os status:\n%v", err)
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
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
	b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
}
