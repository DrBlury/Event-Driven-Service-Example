package observability

import (
	"context"
	"log/slog"

	"github.com/samber/oops"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	LogKeyTraceID = "trace_id"
	LogKeySpanID  = "span_id"
)

func Logger(ctx context.Context, logger *slog.Logger) *slog.Logger {
	log := logger
	if log == nil {
		log = slog.Default()
	}

	traceID, spanID := TraceAndSpanIDs(ctx)
	if traceID == "" {
		return log
	}

	return log.With(
		LogKeyTraceID, traceID,
		LogKeySpanID, spanID,
	)
}

func TraceAndSpanIDs(ctx context.Context) (string, string) {
	if ctx == nil {
		return "", ""
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return "", ""
	}

	return spanContext.TraceID().String(), spanContext.SpanID().String()
}

func Builder(ctx context.Context, domain string, code any) oops.OopsErrorBuilder {
	builder := oops.
		In(domain).
		Code(code)

	traceID, spanID := TraceAndSpanIDs(ctx)
	if traceID != "" {
		builder = builder.Trace(traceID)
	}
	if spanID != "" {
		builder = builder.Span(spanID)
	}

	return builder
}

func RecordError(span trace.Span, err error, attrs ...attribute.KeyValue) {
	if err == nil || span == nil {
		return
	}

	span.RecordError(err, trace.WithAttributes(attrs...))
	span.SetStatus(codes.Error, err.Error())
}
