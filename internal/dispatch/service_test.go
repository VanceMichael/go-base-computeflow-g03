package dispatch_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/computeflow/internal/dispatch"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/incident"
	"github.com/VanceMichael/computeflow/internal/testsupport"
	"github.com/google/uuid"
	"testing"
)

func TestResponderClaimCreatesSingleAssignment(t *testing.T) {
	f := testsupport.New(t)
	r := f.Responder("Medic A")
	i, err := incident.New(f.Store).Open(context.Background(), f.Port.ID, "", "passenger", "p", "high", "unwell", f.Now)
	if err != nil {
		t.Fatal(err)
	}
	if err := dispatch.New(f.Store).Claim(context.Background(), i.ID, r.ID, "dispatcher"); err != nil {
		t.Fatal(err)
	}
	if err := dispatch.New(f.Store).Claim(context.Background(), uuid.NewString(), r.ID, "other"); err == nil {
		t.Fatal("busy responder claimed")
	}
}
func TestResponderReleaseRequiresCurrentVersion(t *testing.T) {
	f := testsupport.New(t)
	r := f.Responder("Medic")
	if _, err := f.Store.DB.Exec(`UPDATE responders SET state='busy' WHERE id=?`, r.ID); err != nil {
		t.Fatal(err)
	}
	if err := dispatch.New(f.Store).Release(context.Background(), r.ID, 99); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("got %v", err)
	}
}
