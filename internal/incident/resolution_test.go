package incident_test

import (
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/incident"
	"testing"
)

func TestResolutionNeedsEvidence(t *testing.T) {
	r := incident.Resolution{ResponderID: "r", Summary: "fixed", ResolvedAt: "now"}
	if err := r.Validate(); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("got %v", err)
	}
}
func TestMedicalResolutionNeedsMedicalEvidence(t *testing.T) {
	triage := incident.Classify("passenger unwell")
	if err := incident.ValidateResolution(triage, "transported by medical team"); err != nil {
		t.Fatal(err)
	}
}
func TestTriageDetectsTrafficIncident(t *testing.T) {
	got := incident.Classify("vehicle stopped in lane")
	if !got.RequiresTrafficControl || got.Severity != "medium" {
		t.Fatalf("%+v", got)
	}
}
func TestNormalizeSummaryCollapsesWhitespace(t *testing.T) {
	if incident.NormalizeSummary("  all   clear ") != "all clear" {
		t.Fatal("not normalized")
	}
}
