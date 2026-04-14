package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var ErrUnknownStatus = fmt.Errorf("unknown task status")

type TaskStatus int
type TaskID int

var statusName = map[TaskStatus]string{
	StatusBacklog:    "backlog",
	StatusInProgress: "in_progress",
	StatusReview:     "review",
	StatusDone:       "done",
	StatusCancelled:  "cancelled",
}

func FromString(str string) (TaskStatus, error) {
	str = strings.ToLower(strings.TrimSpace(str))

	switch str {
	case "backlog":
		return StatusBacklog, nil
	case "in_progress":
		return StatusInProgress, nil
	case "review":
		return StatusReview, nil
	case "done":
		return StatusDone, nil
	case "cancelled":
		return StatusCancelled, nil
	default:
		return StatusBacklog, fmt.Errorf("%w: %s", ErrUnknownStatus, str)
	}
}

func (ts TaskStatus) ToString() string {
	return statusName[ts]
}

func (ts TaskStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(statusName[ts])
}

func (ts *TaskStatus) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "backlog":
		*ts = StatusBacklog
	case "in_progress":
		*ts = StatusInProgress
	case "review":
		*ts = StatusReview
	case "done":
		*ts = StatusDone
	case "cancelled":
		*ts = StatusCancelled
	default:
		return fmt.Errorf("unknown status: %s", s)
	}

	return nil
}

func (ti TaskID) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(ti))
}

func (ti *TaskID) UnmarshalJSON(data []byte) error {
	var res int
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	*ti = TaskID(res)

	return nil
}

type Task struct {
	ID          TaskID     `json:"id"`
	Title       string     `json:"title"`
	Author      string     `json:"author"`
	Description string     `json:"description"`
	Executor    string     `json:"executor"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
