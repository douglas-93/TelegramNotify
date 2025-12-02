package schedule

import (
	"encoding/json"
	"os"
	"sync"
)

const storageFile = "schedules.json"

type Storage struct {
	mu   sync.Mutex
	Jobs map[int64]Job
}

func NewStorage() *Storage {
	return &Storage{
		Jobs: make(map[int64]Job),
	}
}

func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(storageFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.Jobs)
}

func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.Jobs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(storageFile, data, 0644)
}

func (s *Storage) Add(j Job) error {
	s.Jobs[j.ID] = j
	return s.Save()
}

func (s *Storage) Delete(id int64) error {
	delete(s.Jobs, id)
	return s.Save()
}

func (s *Storage) All() []Job {
	list := []Job{}
	for _, j := range s.Jobs {
		list = append(list, j)
	}
	return list
}
