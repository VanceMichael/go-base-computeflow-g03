package risk

import (
	"context"
	"errors"
	"github.com/VanceMichael/harborflow/internal/domain"
	"time"
)

type Verifier interface {
	Verify(context.Context, string) error
}
type StaticVerifier struct {
	Err   error
	Delay time.Duration
}

func (v StaticVerifier) Verify(ctx context.Context, _ string) error {
	if v.Delay > 0 {
		t := time.NewTimer(v.Delay)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
	return v.Err
}

type Service struct{ Verifier Verifier }

func New(v Verifier) *Service { return &Service{Verifier: v} }
func (s *Service) Assess(ctx context.Context, subject string) (string, error) {
	if err := s.Verifier.Verify(ctx, subject); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", domain.ErrUnavailable
	}
	return "cleared", nil
}
