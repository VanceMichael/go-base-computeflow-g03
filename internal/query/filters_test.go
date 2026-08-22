package query_test

import (
	"github.com/VanceMichael/harborflow/internal/query"
	"testing"
	"time"
)

func TestAuditFilterBuildsTenantBoundQuery(t *testing.T) {
	f := query.AuditFilter{PortID: "port", RunID: "run", SubjectType: "passenger", From: time.Unix(1, 0), To: time.Unix(2, 0)}
	if err := f.Validate(); err != nil {
		t.Fatal(err)
	}
	where, args := f.Where(0)
	if where == "" || len(args) != 6 {
		t.Fatalf("%q %v", where, args)
	}
}
func TestAuditFilterRejectsReverseWindow(t *testing.T) {
	f := query.AuditFilter{PortID: "p", From: time.Unix(2, 0), To: time.Unix(1, 0)}
	if err := f.Validate(); err == nil {
		t.Fatal("reverse range accepted")
	}
}
func TestVehicleFilterRequiresPort(t *testing.T) {
	if err := (&query.VehicleFilter{}).Validate(); err == nil {
		t.Fatal("missing port accepted")
	}
}
