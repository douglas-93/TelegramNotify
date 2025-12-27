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
	// Envia mensagem inicial
	processingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ Consultando status dos hosts no Zabbix...")
	tempMsg, _ := b.API.Send(processingMsg)

	hosts, err := monitor.CheckHostsStatus(b.Zabbix)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao consultar Zabbix:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
		log.Println(err)
		return
	}

	msg := "🚥🚥🚥 Status dos Hosts 🚥🚥🚥\n\n"
	for _, h := range hosts {
		msg += h + "\n"
	}

	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, msg)
	b.API.Send(edit)
}

func (b *Bot) handlePrinterCounter(update tgbotapi.Update) {
	// Envia mensagem inicial
	processingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ Coletando contadores das impressoras...")
	tempMsg, _ := b.API.Send(processingMsg)

	printers, err := monitor.GetPrintersCounter(b.Zabbix)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao consultar Zabbix:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
		log.Println(err)
		return
	}

	// Atualiza mensagem
	updateMsg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, "📊 Processando dados...")
	b.API.Send(updateMsg)

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

	// Envia resultado
	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, msg)
	b.API.Send(edit)

	// Atualiza para gerar planilha
	updateMsg = tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, msg+"\n📄 Gerando planilha Excel...")
	b.API.Send(updateMsg)

	excelFile, err := file_handler.GenerateSheet(printers)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao gerar planilha:\n%v", err)
		b.API.Send(tgbotapi.NewMessage(update.Message.Chat.ID, errorMsg))
		log.Println(err)
		return
	}

	// Envia planilha
	b.API.Send(tgbotapi.NewDocument(update.Message.Chat.ID, tgbotapi.FilePath(excelFile)))
	os.Remove(excelFile)

	// Atualiza mensagem final
	finalMsg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, msg+"\n✅ Planilha enviada com sucesso!")
	b.API.Send(finalMsg)
}

func (b *Bot) handleListIp(update tgbotapi.Update) {
	// Envia mensagem inicial
	processingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ Consultando lista de IPs no Zabbix...")
	tempMsg, _ := b.API.Send(processingMsg)

	hostsList, err := b.Zabbix.ListIps()
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao listar Zabbix:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
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

	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, msg)
	b.API.Send(edit)
}

func (b *Bot) handleProtheusStatus(update tgbotapi.Update) {
	// Envia mensagem inicial
	processingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ Consultando status dos serviços Protheus...")
	tempMsg, _ := b.API.Send(processingMsg)

	services, err := b.Zabbix.GetProtheusServiceStatus()
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao pegar os status:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
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

	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, msg)
	b.API.Send(edit)
}
