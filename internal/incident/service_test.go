package incident_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/incident"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"testing"
)

func TestIncidentLifecyclePersistsEachTransition(t *testing.T) {
	f := testsupport.New(t)
	s := incident.New(f.Store)
	i, err := s.Open(context.Background(), f.Port.ID, "", "passenger", "p", "high", "unwell", f.Now)
	if err != nil {
		t.Fatal(err)
	}
	for _, to := range []domain.IncidentState{domain.IncidentAcknowledged, domain.IncidentResolved, domain.IncidentClosed} {
		if err := s.Transition(context.Background(), i, to); err != nil {
			t.Fatal(err)
		}
		i.State = to
		i.Version++
	}
}
func TestIncidentTransitionRejectsSkip(t *testing.T) {
	f := testsupport.New(t)
	i, err := incident.New(f.Store).Open(context.Background(), f.Port.ID, "", "passenger", "p", "high", "unwell", f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := incident.New(f.Store).Transition(context.Background(), i, domain.IncidentClosed); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestIncidentTransitionRejectsStaleVersion(t *testing.T) {
	f := testsupport.New(t)
	i, err := incident.New(f.Store).Open(context.Background(), f.Port.ID, "", "passenger", "p", "high", "unwell", f.Now)
	if err != nil {
		t.Fatal(err)
	}
	i.Version = 4
	if err := incident.New(f.Store).Transition(context.Background(), i, domain.IncidentAcknowledged); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
func TestEvidenceArchiveClosesEveryWriter(t *testing.T) {
	calls := 0
	arch := incident.Archiver{Open: func(context.Context, string) (incident.EvidenceWriter, error) {
		return &writer{closeFn: func() { calls++ }}, nil
	}}
	if err := arch.ArchiveEvidence(context.Background(), []string{"a", "b", "c"}, []byte("evidence")); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("closed=%d", calls)
	}
}
func TestEvidenceArchiveStopsAfterCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	arch := incident.Archiver{Open: func(context.Context, string) (incident.EvidenceWriter, error) {
		t.Fatal("writer should not open")
		return nil, nil
	}}
	if err := arch.ArchiveEvidence(ctx, []string{"a"}, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v", err)
	}
}

type writer struct{ closeFn func() }

func (w *writer) Write(p []byte) (int, error) { return len(p), nil }
func (w *writer) Close() error                { w.closeFn(); return nil }
