package types

import "time"

type ScheduleType string

const (
	ScheduleCron  ScheduleType = "cron"
	ScheduleEvery ScheduleType = "every"
	ScheduleAt    ScheduleType = "at"
)

type Schedule struct {
	Type       ScheduleType `json:"type"`
	Expression string       `json:"expression"`
}

type Job struct {
	Name        string            `json:"name"`
	Command     string            `json:"command"`
	Workdir     string            `json:"workdir"`
	Schedule    Schedule          `json:"schedule"`
	Env         map[string]string `json:"env"`
	Description string            `json:"description,omitempty"`
	RunOnLoad   bool              `json:"run_on_load"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Store struct {
	Version int            `json:"version"`
	Jobs    map[string]Job `json:"jobs"`
}
