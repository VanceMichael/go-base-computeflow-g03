package risk_test

import (
	"context"
	"errors"
	"github.com/VanceMichael/computeflow/internal/domain"
	"github.com/VanceMichael/computeflow/internal/risk"
	"testing"
	"time"
)

func TestRiskAssessmentReturnsClearedOnAcceptedDependency(t *testing.T) {
	state, err := risk.New(risk.StaticVerifier{}).Assess(context.Background(), "passenger")
	if err != nil || state != "cleared" {
		t.Fatalf("%q %v", state, err)
	}
}
func TestRiskAssessmentClassifiesDependencyFailure(t *testing.T) {
	state, err := risk.New(risk.StaticVerifier{Err: errors.New("remote")}).Assess(context.Background(), "passenger")
	if state != "" || !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("%q %v", state, err)
	}
}
func TestRiskAssessmentPreservesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	state, err := risk.New(risk.StaticVerifier{Delay: time.Second}).Assess(ctx, "passenger")
	if state != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("%q %v", state, err)
	}
}
