package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound     = errors.New("computeflow: not found")
	ErrConflict     = errors.New("computeflow: conflict")
	ErrInvalid      = errors.New("computeflow: invalid operation")
	ErrUnauthorized = errors.New("computeflow: unauthorized")
	ErrUnavailable  = errors.New("computeflow: dependency unavailable")
)

type Role string

const (
	RoleCoordinator Role = "coordinator"
	RoleInspector   Role = "inspector"
	RoleDispatcher  Role = "dispatcher"
	RoleAuditor     Role = "auditor"
)

type RunState string

const (
	RunDraft     RunState = "draft"
	RunRunning   RunState = "running"
	RunPaused    RunState = "paused"
	RunCompleted RunState = "completed"
)

type WaveState string

const (
	WavePlanned   WaveState = "planned"
	WaveReleasing WaveState = "releasing"
	WaveReleased  WaveState = "released"
	WaveClosed    WaveState = "closed"
)

type PassengerState string

const (
	PassengerWaiting  PassengerState = "waiting"
	PassengerChecking PassengerState = "checking"
	PassengerCleared  PassengerState = "cleared"
	PassengerHeld     PassengerState = "held"
)

type ScanState string

const (
	ScanPending  ScanState = "pending"
	ScanLeased   ScanState = "leased"
	ScanCleared  ScanState = "cleared"
	ScanRejected ScanState = "rejected"
)

type VehicleState string

const (
	VehicleQueued    VehicleState = "queued"
	VehicleAssessing VehicleState = "assessing"
	VehicleAdmitted  VehicleState = "admitted"
	VehicleHeld      VehicleState = "held"
	VehicleInspected VehicleState = "inspected"
)

type LaneState string

const (
	LaneOpen    LaneState = "open"
	LaneClosing LaneState = "closing"
	LaneClosed  LaneState = "closed"
)

type IncidentState string

const (
	IncidentOpen         IncidentState = "open"
	IncidentAcknowledged IncidentState = "acknowledged"
	IncidentResolved     IncidentState = "resolved"
	IncidentClosed       IncidentState = "closed"
)

type OutboxState string

const (
	OutboxPending   OutboxState = "pending"
	OutboxLeased    OutboxState = "leased"
	OutboxDelivered OutboxState = "delivered"
	OutboxDead      OutboxState = "dead"
)

type Port struct {
	ID, Code, Name, Timezone string
	CreatedAt                time.Time
}
type User struct {
	ID, PortID, Email, DisplayName string
	Role                           Role
	Active                         bool
	CreatedAt                      time.Time
}
type Session struct {
	ID, UserID, TokenHash string
	ExpiresAt             time.Time
	RevokedAt             *time.Time
	CreatedAt             time.Time
}
type StressRun struct {
	ID, PortID, Name string
	State            RunState
	Version          int
	StartsAt         time.Time
	EndsAt           *time.Time
	CreatedBy        string
	CreatedAt        time.Time
}
type PassengerWave struct {
	ID, RunID           string
	SequenceNo, Version int
	State               WaveState
	PlannedAt           time.Time
	ReleasedAt          *time.Time
}
type Passenger struct {
	ID, WaveID, DocumentKey string
	State                   PassengerState
	Version                 int
	CreatedAt               time.Time
}
type Gate struct {
	ID, PortID string
	GateNo     int
	Mode       string
	Active     bool
}
type GateScan struct {
	ID, PassengerID, GateID, LeaseOwner, LeaseToken string
	Stage, Version                                  int
	State                                           ScanState
	LeaseUntil                                      *time.Time
	CreatedAt                                       time.Time
}
type VehicleBatch struct {
	ID, RunID, ManifestKey string
	State                  string
	Version                int
	CreatedAt              time.Time
}
type Vehicle struct {
	ID, BatchID, PlateKey string
	State                 VehicleState
	Version               int
	CreatedAt             time.Time
}
type Lane struct {
	ID, PortID string
	LaneNo     int
	State      LaneState
	Version    int
}
type Incident struct {
	ID, PortID, RunID, SubjectType, SubjectID, Severity, Description string
	State                                                            IncidentState
	Version                                                          int
	CreatedAt                                                        time.Time
}
type Responder struct {
	ID, PortID, Name, State string
	Version                 int
}
type OutboxMessage struct {
	ID, PortID, EventKey, Payload, State, Owner, IdempotencyKey string
	LeaseUntil                                                  *time.Time
	Attempts                                                    int
	LastError                                                   string
	CreatedAt                                                   time.Time
	DeliveredAt                                                 *time.Time
}
type CapacitySnapshot struct {
	ID, RunID                           string
	WindowStart, WindowEnd              time.Time
	Passengers, Cleared, Held, Vehicles int
	State                               string
	CreatedAt                           time.Time
}
type AuditEvent struct {
	ID, PortID, ActorID, Action, SubjectType, SubjectID, Outcome, RequestID, Details string
	CreatedAt                                                                        time.Time
}

func TransitionRun(from, to RunState) error {
	if from == RunDraft && to == RunRunning || from == RunRunning && (to == RunPaused || to == RunCompleted) || from == RunPaused && (to == RunRunning || to == RunCompleted) {
		return nil
	}
	return fmt.Errorf("%w: run %s -> %s", ErrInvalid, from, to)
}
func TransitionWave(from, to WaveState) error {
	if from == WavePlanned && to == WaveReleasing || from == WaveReleasing && (to == WaveReleased || to == WavePlanned) || from == WaveReleased && to == WaveClosed {
		return nil
	}
	return fmt.Errorf("%w: wave %s -> %s", ErrInvalid, from, to)
}
func TransitionIncident(from, to IncidentState) error {
	if from == IncidentOpen && to == IncidentAcknowledged || from == IncidentAcknowledged && to == IncidentResolved || from == IncidentResolved && to == IncidentClosed {
		return nil
	}
	return fmt.Errorf("%w: incident %s -> %s", ErrInvalid, from, to)
}
