package request

import "context"

type key int

const (
	requestIDKey key = iota
	actorKey
	portKey
)

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
func RequestID(ctx context.Context) string { v, _ := ctx.Value(requestIDKey).(string); return v }
func WithActor(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, actorKey, id)
}
func Actor(ctx context.Context) string { v, _ := ctx.Value(actorKey).(string); return v }
func WithPort(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, portKey, id)
}
func Port(ctx context.Context) string { v, _ := ctx.Value(portKey).(string); return v }
