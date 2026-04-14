package validator

import (
	"strings"
	"tracker/internal/model"
)

func ValidateTask(task *model.Task) error {
	if strings.TrimSpace(task.Title) == "" {
		return model.ErrUnknownStatus
	}

	if len(task.Title) > model.TitleMaxLength {
		return model.ErrUnknownStatus
	}

	return nil
}
