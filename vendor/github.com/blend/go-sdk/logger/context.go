package logger

import (
	"context"
	"time"
)

type loggerKey struct{}

// WithLogger adds the logger to a context.
func WithLogger(ctx context.Context, log Log) context.Context {
	return context.WithValue(ctx, loggerKey{}, log)
}

// GetLogger gets a logger off a context.
func GetLogger(ctx context.Context) Log {
	if value := ctx.Value(loggerKey{}); value != nil {
		if typed, ok := value.(Log); ok {
			return typed
		}
	}
	return nil
}

type triggerTimestampKey struct{}

// WithTriggerTimestamp returns a new context with a given timestamp value.
// It is used by the scope to connote when an event was triggered.
func WithTriggerTimestamp(ctx context.Context, ts time.Time) context.Context {
	return context.WithValue(ctx, triggerTimestampKey{}, ts)
}

// GetTriggerTimestamp gets when an event was triggered off a context.
func GetTriggerTimestamp(ctx context.Context) time.Time {
	if raw := ctx.Value(triggerTimestampKey{}); raw != nil {
		if typed, ok := raw.(time.Time); ok {
			return typed
		}
	}
	return time.Time{}
}

type timestampKey struct{}

// WithTimestamp returns a new context with a given timestamp value.
func WithTimestamp(ctx context.Context, ts time.Time) context.Context {
	return context.WithValue(ctx, timestampKey{}, ts)
}

// GetTimestamp gets a timestampoff a context.
func GetTimestamp(ctx context.Context) time.Time {
	if raw := ctx.Value(timestampKey{}); raw != nil {
		if typed, ok := raw.(time.Time); ok {
			return typed
		}
	}
	return time.Time{}
}

type pathKey struct{}

// WithPath returns a new context with a given additional path segment(s).
func WithPath(ctx context.Context, path ...string) context.Context {
	return context.WithValue(ctx, pathKey{}, path)
}

// GetPath gets a path off a context.
func GetPath(ctx context.Context) []string {
	if raw := ctx.Value(pathKey{}); raw != nil {
		if typed, ok := raw.([]string); ok {
			return typed
		}
	}
	return nil
}

type labelsKey struct{}

// WithLabels returns a new context with a given additional labels.
func WithLabels(ctx context.Context, labels Labels) context.Context {
	return context.WithValue(ctx, labelsKey{}, labels)
}

// WithLabel returns a new context with a given additional label.
func WithLabel(ctx context.Context, key, value string) context.Context {
	existing := GetLabels(ctx)
	if existing == nil {
		existing = make(Labels)
	}
	existing[key] = value
	return context.WithValue(ctx, labelsKey{}, existing)
}

// GetLabels gets labels off a context.
func GetLabels(ctx context.Context) Labels {
	if raw := ctx.Value(labelsKey{}); raw != nil {
		if typed, ok := raw.(Labels); ok {
			return typed
		}
	}
	return nil
}

type annotationsKey struct{}

// WithAnnotations returns a new context with a given additional annotations.
func WithAnnotations(ctx context.Context, annotations Annotations) context.Context {
	return context.WithValue(ctx, annotationsKey{}, annotations)
}

// WithAnnotation returns a new context with a given additional annotation.
func WithAnnotation(ctx context.Context, key, value string) context.Context {
	existing := GetAnnotations(ctx)
	if existing == nil {
		existing = make(Annotations)
	}
	existing[key] = value
	return context.WithValue(ctx, annotationsKey{}, existing)
}

// GetAnnotations gets annotations off a context.
func GetAnnotations(ctx context.Context) Annotations {
	if raw := ctx.Value(annotationsKey{}); raw != nil {
		if typed, ok := raw.(Annotations); ok {
			return typed
		}
	}
	return nil
}

type skipTriggerKey struct{}

type skipWriteKey struct{}

// WithSkipTrigger sets the context to skip logger listener triggers.
// The event will still be written unless you also use `WithSkipWrite`.
func WithSkipTrigger(ctx context.Context, skipTrigger bool) context.Context {
	return context.WithValue(ctx, skipTriggerKey{}, skipTrigger)
}

// WithSkipWrite sets the context to skip writing the event to the output stream.
// The event will still trigger listeners unless you also use `WithSkipTrigger`.
func WithSkipWrite(ctx context.Context, skipWrite bool) context.Context {
	return context.WithValue(ctx, skipWriteKey{}, skipWrite)
}

// IsSkipTrigger returns if we should skip triggering logger listeners for a context.
func IsSkipTrigger(ctx context.Context) bool {
	if v := ctx.Value(skipTriggerKey{}); v != nil {
		if typed, ok := v.(bool); ok {
			return typed
		}
	}
	return false
}

// IsSkipWrite returns if we should skip writing to the event stream for a context.
func IsSkipWrite(ctx context.Context) bool {
	if v := ctx.Value(skipWriteKey{}); v != nil {
		if typed, ok := v.(bool); ok {
			return typed
		}
	}
	return false
}
