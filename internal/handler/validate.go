package handler

import (
	"strings"

	"github.com/team-task-api/internal/model"
)

func isValidEmail(email string) bool {
	at := strings.Index(email, "@")
	if at < 1 {
		return false
	}
	dot := strings.LastIndex(email[at:], ".")
	return dot > 1 && dot < len(email[at:])-1
}

func isValidTaskStatus(status string) bool {
	switch model.TaskStatus(status) {
	case model.StatusPending, model.StatusInProgress, model.StatusDone:
		return true
	}
	return false
}
