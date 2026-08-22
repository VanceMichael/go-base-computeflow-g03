package request_test

import (
	"context"
	"github.com/VanceMichael/harborflow/internal/request"
	"testing"
)

func TestRequestContextCarriesOperationalIdentity(t *testing.T) {
	ctx := request.WithRequestID(context.Background(), "req")
	ctx = request.WithActor(ctx, "actor")
	ctx = request.WithPort(ctx, "port")
	if request.RequestID(ctx) != "req" || request.Actor(ctx) != "actor" || request.Port(ctx) != "port" {
		t.Fatal("context values lost")
	}
}
func TestRequestContextReturnsEmptyForUnknownValues(t *testing.T) {
	if request.RequestID(context.Background()) != "" || request.Actor(context.Background()) != "" || request.Port(context.Background()) != "" {
		t.Fatal("unexpected context value")
	}
}
