package schedule

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
)

type Manager struct {
	Sched *gocron.Scheduler
	Jobs  map[int64]*gocron.Job
}

func NewManager() *Manager {
	s := gocron.NewScheduler(time.UTC)
	return &Manager{
		Sched: s,
		Jobs:  make(map[int64]*gocron.Job),
	}
}

func (m *Manager) Start() {
	m.Sched.StartAsync()
}

func (m *Manager) Add(j Job, executor func()) error {
	job, err := m.Sched.Cron(j.Cron).Do(executor)
	if err != nil {
		return fmt.Errorf("erro ao criar cron: %v", err)
	}
	m.Jobs[j.ID] = job
	return nil
}

func (m *Manager) Remove(id int64) {
	if job, ok := m.Jobs[id]; ok {
		m.Sched.RemoveByReference(job)
		delete(m.Jobs, id)
	}
}
