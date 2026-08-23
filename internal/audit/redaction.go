package audit

import (
	"github.com/VanceMichael/computeflow/internal/domain"
	"strings"
)

var sensitiveKeys = map[string]bool{"document_key": true, "face_template": true, "fingerprint": true, "token": true}

func Redact(details map[string]string) map[string]string {
	out := map[string]string{}
	for key, value := range details {
		if sensitiveKeys[strings.ToLower(key)] {
			out[key] = "[redacted]"
		} else {
			out[key] = value
		}
	}
	return out
}
func PublicEvent(e domain.AuditEvent) domain.AuditEvent {
	e.ActorID = ""
	e.Details = "redacted"
	return e
}
func IsSensitive(key string) bool { return sensitiveKeys[strings.ToLower(key)] }
