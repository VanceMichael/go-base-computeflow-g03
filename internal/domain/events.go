package domain

import "time"

type EventType string

const (
	EventWaveReleased     EventType = "wave.released"
	EventGateCleared      EventType = "gate.cleared"
	EventVehicleHeld      EventType = "vehicle.held"
	EventIncidentOpened   EventType = "incident.opened"
	EventCapacityRecorded EventType = "capacity.recorded"
)

type OperationalEvent struct {
	ID          string
	PortID      string
	Type        EventType
	SubjectType string
	SubjectID   string
	ActorID     string
	RequestID   string
	OccurredAt  time.Time
	Payload     map[string]string
}

func (e OperationalEvent) IsScoped() bool {
	return e.ID != "" && e.PortID != "" && e.SubjectID != "" && e.Type != ""
}
func (e OperationalEvent) AsAuditDetails() string {
	if len(e.Payload) == 0 {
		return "{}"
	}
	out := "{"
	first := true
	for k, v := range e.Payload {
		if !first {
			out += ","
		}
		out += "\"" + k + "\":\"" + v + "\""
		first = false
	}
	return out + "}"
}
