package flow_test

import (
	"github.com/VanceMichael/computeflow/internal/flow"
	"testing"
)

func TestReleaseCommandRequiresAllIdentityFields(t *testing.T) {
	if err := (flow.ReleaseCommand{Documents: []string{"D"}}).Validate(); err == nil {
		t.Fatal("incomplete command accepted")
	}
}
func TestReleaseCommandRejectsDuplicateDocuments(t *testing.T) {
	c := flow.ReleaseCommand{RunID: "r", WaveID: "w", ActorID: "a", RequestID: "q", Documents: []string{"D", "D"}}
	if err := c.Validate(); err == nil {
		t.Fatal("duplicate accepted")
	}
}
func TestReleaseCommandTrimsDocumentKeys(t *testing.T) {
	c := flow.ReleaseCommand{RunID: "r", WaveID: "w", ActorID: "a", RequestID: "q", Documents: []string{" D "}}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
}
func TestPauseCommandRequiresReason(t *testing.T) {
	if err := (flow.PauseCommand{WaveID: "w", ActorID: "a"}).Validate(); err == nil {
		t.Fatal("missing reason accepted")
	}
}
func TestNormalizeDocumentsDropsBlankEntries(t *testing.T) {
	got := flow.NormalizeDocuments([]string{" A ", " ", "B"})
	if len(got) != 2 || got[0] != "A" {
		t.Fatalf("%v", got)
	}
}
