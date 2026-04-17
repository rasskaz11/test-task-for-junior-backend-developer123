package task

import "time"

type Status string

const (
    StatusNew        Status = "new"
    StatusInProgress Status = "in_progress"
    StatusDone       Status = "done"
)

type Task struct {
    ID          int64           `json:"id"`
    Title       string          `json:"title"`
    Description string          `json:"description"`
    Status      Status          `json:"status"`
    CreatedAt   time.Time       `json:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at"`
    Recurrence  *TaskRecurrence `json:"recurrence,omitempty"` 
}

type TaskRecurrence struct {
    ID       int    `json:"id"`
    TaskID   int    `json:"task_id"`
    Type     string `json:"type"`
    Value    string `json:"value"`
    Interval int    `json:"interval"`
}

func (s Status) Valid() bool {
    switch s {
    case StatusNew, StatusInProgress, StatusDone:
        return true
    default:
        return false
    }
}

var ErrNotFound = "task not found"