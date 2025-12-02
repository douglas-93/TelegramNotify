package schedule

type Job struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Cron    string `json:"cron"`
	Command string `json:"command"`
	ChatID  int64  `json:"chat_id"`
}
