package domain

import (
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string { return fmt.Sprintf("%s: %s", e.Field, e.Message) }
func ValidatePortCode(v string) error {
	v = strings.TrimSpace(v)
	if len(v) < 3 || len(v) > 16 {
		return ValidationError{"port_code", "must contain 3-16 characters"}
	}
	return nil
}
func ValidateDocumentKey(v string) error {
	if strings.TrimSpace(v) == "" || len(v) > 64 {
		return ValidationError{"document_key", "must be non-empty and at most 64 characters"}
	}
	return nil
}
func ValidateManifestKey(v string) error {
	if strings.TrimSpace(v) == "" || len(v) > 80 {
		return ValidationError{"manifest_key", "must be non-empty and at most 80 characters"}
	}
	return nil
}
func ValidateSequence(v int) error {
	if v < 1 {
		return ValidationError{"sequence", "must be positive"}
	}
	return nil
}
func ValidateWindow(from, to int64) error {
	if from >= to {
		return ValidationError{"window", "start must precede end"}
	}
	return nil
}
func IsTerminalRun(s RunState) bool           { return s == RunCompleted }
func IsTerminalIncident(s IncidentState) bool { return s == IncidentClosed }
func IsActivePassenger(s PassengerState) bool { return s == PassengerWaiting || s == PassengerChecking }
