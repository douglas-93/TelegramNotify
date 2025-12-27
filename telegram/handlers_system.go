package bot

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

func (b *Bot) handlePing(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o IP. Ex: /ping 192.168.0.1")
		b.API.Send(msg)
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
		b.API.Send(msg)
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

func (b *Bot) handleRestartWindowsHost(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o hostname. Ex: /restart_win \\\\LVMAQUINA")
		b.API.Send(msg)
		return
	}

	host := parts[1]
	log.Printf("Handler restart_win acionado, destino: %s", host)

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
		b.API.Send(msg)
		return
	}
	m := fmt.Sprintf("✅ Comando executado para: %s", host)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
	b.API.Send(msg)
}

func (b *Bot) handleShutdownWindowsHost(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) <= 1 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Informe o hostname. Ex: /shutdown_win \\\\LVMAQUINA")
		b.API.Send(msg)
		return
	}
	host := parts[1]
	log.Printf("Handler shutdown_win acionado, destino: %s", host)

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
		b.API.Send(msg)
		return
	}
	m := fmt.Sprintf("✅ Comando executado para: %s", host)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, m)
	b.API.Send(msg)
}

func (b *Bot) handleRemoteServices(update tgbotapi.Update) {
	parts := strings.Split(update.Message.Text, " ")
	if len(parts) < 4 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Uso: /services <IP/Host> <start|stop|restart> <serviço1> [serviço2] ...")
		b.API.Send(msg)
		return
	}

	host := parts[1]
	operation := strings.ToLower(parts[2])
	services := parts[3:]

	var op ServiceOperation
	switch operation {
	case "start":
		op = OperationStart
	case "stop":
		op = OperationStop
	case "restart":
		op = OperationRestart
	default:
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Operação inválida. Use: start, stop ou restart")
		b.API.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⏳ Conectando em %s...", host))
	tempMsg, _ := b.API.Send(msg)

	// Conectando ao gerenciador de serviços remoto
	m, err := mgr.ConnectRemote(host)
	if err != nil {
		text := fmt.Sprintf("❌ Erro ao conectar no host %s: %v", host, err)
		edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, text)
		b.API.Send(edit)
		return
	}
	defer m.Disconnect()

	// Atualiza mensagem informando execução
	text := fmt.Sprintf("⚙️ Executando %s em %d serviço(s) no host %s...", operation, len(services), host)
	edit := tgbotapi.NewEditMessageText(update.Message.Chat.ID, tempMsg.MessageID, text)
	b.API.Send(edit)

	executeOperation(b, update.Message.Chat.ID, m, services, op)
}

// ServiceOperation define o tipo de operação a ser realizada
type ServiceOperation string

const (
	OperationStop    ServiceOperation = "stop"
	OperationStart   ServiceOperation = "start"
	OperationRestart ServiceOperation = "restart"
)

// ServiceResult representa o resultado de uma operação em um serviço
type ServiceResult struct {
	ServiceName string
	Operation   ServiceOperation
	Success     bool
	Error       error
}

// executeOperation executa a operação especificada nos serviços
func executeOperation(b *Bot, chatID int64, m *mgr.Mgr, services []string, operation ServiceOperation) {
	var results []ServiceResult

	switch operation {
	case OperationStop:
		results = stopServices(m, services)
	case OperationStart:
		results = startServices(m, services)
	case OperationRestart:
		// Stop
		stopResults := stopServices(m, services)
		// Start
		time.Sleep(2 * time.Second)
		startResults := startServices(m, services)

		results = append(stopResults, startResults...)
	}

	sendResults(b, chatID, results)
}

// sendResults envia o relatório para o chat
func sendResults(b *Bot, chatID int64, results []ServiceResult) {
	var sb strings.Builder
	sb.WriteString("📋 *Relatório de Serviços:*\n\n")

	for _, result := range results {
		icon := "✅"
		if !result.Success {
			icon = "❌"
		}

		opText := ""
		switch result.Operation {
		case OperationStop:
			opText = "Parar"
		case OperationStart:
			opText = "Iniciar"
		}

		sb.WriteString(fmt.Sprintf("%s *%s* (%s): ", icon, result.ServiceName, opText))
		if result.Success {
			sb.WriteString("Sucesso")
		} else {
			sb.WriteString(fmt.Sprintf("Erro: %v", result.Error))
		}
		sb.WriteString("\n")
	}

	msg := tgbotapi.NewMessage(chatID, sb.String())
	msg.ParseMode = "Markdown"
	b.API.Send(msg)
}

// stopServices para todos os serviços concorrentemente
func stopServices(m *mgr.Mgr, services []string) []ServiceResult {
	var wg sync.WaitGroup
	results := make(chan ServiceResult, len(services))

	for _, serviceName := range services {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			result := stopService(m, name)
			results <- result
		}(serviceName)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return collectResults(results)
}

// startServices inicia todos os serviços concorrentemente
func startServices(m *mgr.Mgr, services []string) []ServiceResult {
	var wg sync.WaitGroup
	results := make(chan ServiceResult, len(services))

	for _, serviceName := range services {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			result := startService(m, name)
			results <- result
		}(serviceName)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return collectResults(results)
}

// stopService para um serviço específico
func stopService(m *mgr.Mgr, serviceName string) ServiceResult {
	s, err := m.OpenService(serviceName)
	if err != nil {
		return ServiceResult{
			ServiceName: serviceName,
			Operation:   OperationStop,
			Success:     false,
			Error:       fmt.Errorf("abrir serviço: %w", err),
		}
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		// Se já estiver parado, control pode falhar ou retornar erro específico.
		// Vamos checar status atual antes?
		// O snippet original não checava, mas é boa prática.
		// Seguirei o snippet original para fidelidade, mas ajustando msg de erro.
		return ServiceResult{
			ServiceName: serviceName,
			Operation:   OperationStop,
			Success:     false,
			Error:       fmt.Errorf("parar serviço: %w", err),
		}
	}

	timeout := time.Now().Add(30 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return ServiceResult{
				ServiceName: serviceName,
				Operation:   OperationStop,
				Success:     false,
				Error:       fmt.Errorf("timeout ao parar"),
			}
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return ServiceResult{
				ServiceName: serviceName,
				Operation:   OperationStop,
				Success:     false,
				Error:       fmt.Errorf("consultar status: %w", err),
			}
		}
	}

	return ServiceResult{ServiceName: serviceName, Operation: OperationStop, Success: true}
}

// startService inicia um serviço específico
func startService(m *mgr.Mgr, serviceName string) ServiceResult {
	s, err := m.OpenService(serviceName)
	if err != nil {
		return ServiceResult{
			ServiceName: serviceName,
			Operation:   OperationStart,
			Success:     false,
			Error:       fmt.Errorf("abrir serviço: %w", err),
		}
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return ServiceResult{
			ServiceName: serviceName,
			Operation:   OperationStart,
			Success:     false,
			Error:       fmt.Errorf("iniciar serviço: %w", err),
		}
	}

	timeout := time.Now().Add(30 * time.Second)
	for {
		status, err := s.Query()
		if err != nil {
			return ServiceResult{
				ServiceName: serviceName,
				Operation:   OperationStart,
				Success:     false,
				Error:       fmt.Errorf("consultar status: %w", err),
			}
		}

		if status.State == svc.Running {
			break
		}

		if time.Now().After(timeout) {
			return ServiceResult{
				ServiceName: serviceName,
				Operation:   OperationStart,
				Success:     false,
				Error:       fmt.Errorf("timeout ao iniciar"),
			}
		}
		time.Sleep(300 * time.Millisecond)
	}

	return ServiceResult{ServiceName: serviceName, Operation: OperationStart, Success: true}
}

// collectResults coleta todos os resultados do canal
func collectResults(results chan ServiceResult) []ServiceResult {
	collected := make([]ServiceResult, 0)
	for result := range results {
		collected = append(collected, result)
	}
	return collected
}
