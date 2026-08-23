package risk

import (
	"context"
	"fmt"
	"github.com/VanceMichael/computeflow/internal/domain"
)

type Evaluator struct {
	Policy     Policy
	Dependency Verifier
}

func (e Evaluator) Evaluate(ctx context.Context, institution string, confidence float64) (string, string, error) {
	if e.Dependency == nil {
		return "", "", fmt.Errorf("%w: verifier missing", domain.ErrUnavailable)
	}
	if err := e.Dependency.Verify(ctx, institution); err != nil {
		return "", "", err
	}
	return e.Policy.Evaluate(institution, confidence)
}
