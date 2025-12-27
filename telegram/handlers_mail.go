package bot

import (
	"LapaTelegramBot/file_handler"
	"LapaTelegramBot/mailer"
	"LapaTelegramBot/monitor"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleSendMailCounter(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 2 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Uso: /send_mail_counter <email1> [email2] ...")
		b.API.Send(msg)
		return
	}

	emails := parts[1:]

	// Envia mensagem de processamento
	processingMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "⏳ Coletando dados das impressoras...")
	tempMsg, _ := b.API.Send(processingMsg)

	// Obtém contadores das impressoras
	printers, err := monitor.GetPrintersCounter(b.Zabbix)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao consultar Zabbix:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
		log.Println(err)
		return
	}

	// Atualiza mensagem
	updateMsg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, "📊 Gerando planilha...")
	b.API.Send(updateMsg)

	// Gera planilha Excel
	excelFile, err := file_handler.GenerateSheet(printers)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao gerar planilha:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
		log.Printf("Erro ao gerar planilha: %v", err)
		return
	}
	defer os.Remove(excelFile)

	// Atualiza mensagem
	updateMsg = tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, "📧 Enviando email...")
	b.API.Send(updateMsg)

	// Monta corpo do email em HTML
	htmlBody := buildPrinterCounterEmailHTML(printers)

	// Prepara email
	emailMsg := mailer.EmailMessage{
		From:        "telegram.bot@lapavermelha.com.br",
		To:          emails,
		Subject:     "Relatório de Contadores de Impressoras",
		HTMLBody:    htmlBody,
		Attachments: []string{excelFile},
	}

	// Envia email
	err = b.Mailer.SendEmail(emailMsg)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Erro ao enviar email:\n%v", err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, errorMsg)
		b.API.Send(edit)
		log.Printf("Erro ao enviar email: %v", err)
		return
	}

	// Mensagem de sucesso
	successMsg := fmt.Sprintf("✅ Email enviado com sucesso para:\n%s", strings.Join(emails, "\n"))
	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, successMsg)
	b.API.Send(edit)
}

func buildPrinterCounterEmailHTML(printers []monitor.Printer) string {
	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<style>
		body { font-family: Arial, sans-serif; margin: 20px; }
		h1 { color: #333; }
		table { border-collapse: collapse; width: 100%; margin-top: 20px; }
		th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
		th { background-color: #4CAF50; color: white; }
		tr:nth-child(even) { background-color: #f2f2f2; }
		.footer { margin-top: 20px; font-size: 12px; color: #666; }
	</style>
</head>
<body>
	<h1>🔢 Relatório de Contadores de Impressoras</h1>
	<p>Segue abaixo o relatório atualizado dos contadores das impressoras:</p>
	<table>
		<tr>
			<th>Impressora</th>
			<th>Preto e Branco</th>
			<th>Colorido</th>
			<th>Total</th>
		</tr>`)

	for _, printer := range printers {
		sb.WriteString(fmt.Sprintf(`
		<tr>
			<td><strong>%s</strong></td>
			<td>%d</td>
			<td>%d</td>
			<td>%d</td>
		</tr>`, printer.HostData.Host, printer.BlackCounter, printer.ColorCounter, printer.TotalCounter))
	}

	sb.WriteString(`
	</table>
	<div class="footer">
		<p>Relatório gerado automaticamente pelo LapaTelegramBot</p>
		<p>Em anexo você encontra a planilha Excel com os dados detalhados.</p>
	</div>
</body>
</html>`)

	return sb.String()
}
