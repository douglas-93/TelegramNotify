package schedule

import (
	"errors"
	"regexp"
)

var cronRegex = regexp.MustCompile(`^(\S+\s+){4}\S+$`) // 5 campos

func ValidateCron(expr string) error {
	if !cronRegex.MatchString(expr) {
		return errors.New("expressão CRON inválida. Use 5 campos, ex: '0 8 * * *'")
	}
	return nil
}

func CronHelp() string {
	return `
🕒🕒🕒 Como criar agendamentos 🕒🕒🕒

O CRON originalmente é um agendador de tarefas utilizado no sistema Linux. Para agendamento neste bot, você deverá utilizar a notação correspondente.
A expressão CRON é composta por 5 campos, sendo eles:
• Minutos
• Horas
• Dia
• Mês
• Dia da Semana

A expressão deve ser escrita em uma linha, conforme exemplos de uso:

• Todo dia às 14:30
/schedule_add 30 14 * * * /comando

• Toda segunda-feira às 09:00
/schedule_add 0 9 * * 1 /comando

• Primeiro dia de cada mês às 07:00
/schedule_add 0 7 1 * * /comando

• A cada 2 horas
/schedule_add 0 */2 * * * /comando

• De segunda a sexta às 18:00
/schedule_add 0 18 * * 1-5 /comando
`
}
