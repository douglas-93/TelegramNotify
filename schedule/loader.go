package schedule

import "log"

func LoadExistingJobs(s *Storage, m *Manager, handler func(cmd string, chatID int64)) {
	for _, j := range s.All() {

		// cria nova variável para capturar corretamente na closure
		job := j

		m.Add(job, func() {
			log.Printf("Carregando job agendado: %s (ID: %d)", job.Command, job.ID)
			handler(job.Command, job.ChatID)
		})
	}
}
