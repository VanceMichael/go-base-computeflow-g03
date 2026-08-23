package domain_test

import (
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"testing"
)

func TestRunStateMachineAcceptsOperationalPath(t *testing.T) {
	for _, step := range [][2]domain.RunState{{domain.RunDraft, domain.RunRunning}, {domain.RunRunning, domain.RunPaused}, {domain.RunPaused, domain.RunRunning}, {domain.RunRunning, domain.RunCompleted}} {
		if err := domain.TransitionRun(step[0], step[1]); err != nil {
			t.Fatalf("%s -> %s: %v", step[0], step[1], err)
		}
	}
}
func TestRunStateMachineRejectsSkippingPreparation(t *testing.T) {
	if err := domain.TransitionRun(domain.RunDraft, domain.RunCompleted); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}
func TestRunStateMachineRejectsResumingCompletedRun(t *testing.T) {
	if err := domain.TransitionRun(domain.RunCompleted, domain.RunRunning); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}
func TestWaveStateMachineAcceptsReleaseAndClose(t *testing.T) {
	if err := domain.TransitionWave(domain.WavePlanned, domain.WaveReleasing); err != nil {
		t.Fatal(err)
	}
	if err := domain.TransitionWave(domain.WaveReleasing, domain.WaveReleased); err != nil {
		t.Fatal(err)
	}
	if err := domain.TransitionWave(domain.WaveReleased, domain.WaveClosed); err != nil {
		t.Fatal(err)
	}
}
func TestWaveStateMachineAllowsRetryToPlanned(t *testing.T) {
	if err := domain.TransitionWave(domain.WaveReleasing, domain.WavePlanned); err != nil {
		t.Fatal(err)
	}
}
func TestWaveStateMachineRejectsClosingBeforeRelease(t *testing.T) {
	if err := domain.TransitionWave(domain.WavePlanned, domain.WaveClosed); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestIncidentStateMachineRequiresAcknowledgement(t *testing.T) {
	if err := domain.TransitionIncident(domain.IncidentOpen, domain.IncidentResolved); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestIncidentStateMachineAcceptsResponseLifecycle(t *testing.T) {
	for _, step := range [][2]domain.IncidentState{{domain.IncidentOpen, domain.IncidentAcknowledged}, {domain.IncidentAcknowledged, domain.IncidentResolved}, {domain.IncidentResolved, domain.IncidentClosed}} {
		if err := domain.TransitionIncident(step[0], step[1]); err != nil {
			t.Fatal(err)
		}
	}
}
func TestDocumentValidationRejectsBlankValue(t *testing.T) {
	if domain.ValidateDocumentKey(" ") == nil {
		t.Fatal("blank document must fail")
	}
}
func TestDocumentValidationAcceptsOperationalKey(t *testing.T) {
	if err := domain.ValidateDocumentKey("HKID-TEST-001"); err != nil {
		t.Fatal(err)
	}
}
func TestManifestValidationRejectsOversizedValue(t *testing.T) {
	if err := domain.ValidateManifestKey(string(make([]byte, 81))); err == nil {
		t.Fatal("oversized manifest must fail")
	}
}
func TestPortValidationRequiresMinimumLength(t *testing.T) {
	if err := domain.ValidatePortCode("A"); err == nil {
		t.Fatal("short code must fail")
	}
}
func TestSequenceValidationRejectsZero(t *testing.T) {
	if err := domain.ValidateSequence(0); err == nil {
		t.Fatal("zero sequence must fail")
	}
}
func TestWindowValidationRejectsReverseOrder(t *testing.T) {
	if err := domain.ValidateWindow(10, 10); err == nil {
		t.Fatal("equal window must fail")
	}
}
func TestPageValidationAcceptsBoundedPage(t *testing.T) {
	p, err := domain.NewPage(50, 100)
	if err != nil || p.Limit != 50 || p.Offset != 100 {
		t.Fatalf("%+v %v", p, err)
	}
}
func TestPageValidationRejectsLargePage(t *testing.T) {
	if _, err := domain.NewPage(201, 0); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestPageInfoReportsMoreRows(t *testing.T) {
	p, _ := domain.NewPage(20, 20)
	info := domain.NewPageInfo(50, p)
	if !info.HasMore || info.Total != 50 {
		t.Fatalf("%+v", info)
	}
}
func TestTerminalStateHelpers(t *testing.T) {
	if !domain.IsTerminalRun(domain.RunCompleted) {
		t.Fatal("completed run should be terminal")
	}
	if domain.IsTerminalIncident(domain.IncidentOpen) {
		t.Fatal("open incident is active")
	}
}
