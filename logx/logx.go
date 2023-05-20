package logx

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"strings"
)

type ctxKey string

const ctxKeyTraceID ctxKey = "trace_id"

type Severity string

const (
	SeverityInfo  Severity = "info"
	SeverityError Severity = "error"
)

func Infof(ctx context.Context, format string, v ...any) {
	Logf(ctx, SeverityInfo, format, v...)
}

func Errorf(ctx context.Context, format string, v ...any) {
	Logf(ctx, SeverityError, format, v...)
}

func Logf(ctx context.Context, severity Severity, format string, v ...any) {
	prefix := fmt.Sprintf("[%s] [TraceID: %s] ", strings.ToUpper(string(severity)), TraceIdFromContext(ctx))
	msg := fmt.Sprintf(format, v...)
	log.Println(prefix, msg)
}

func Info(ctx context.Context, a ...any) {
	Log(ctx, SeverityInfo, a...)
}

func Error(ctx context.Context, a ...any) {
	Log(ctx, SeverityError, a...)
}

func Log(ctx context.Context, severity Severity, a ...any) {
	prefix := fmt.Sprintf("[%s] [TraceID: %s] ", strings.ToUpper(string(severity)), TraceIdFromContext(ctx))
	args := make([]any, 0, len(a)+1)
	args = append(args, prefix)
	for _, arg := range a {
		args = append(args, arg)
	}
	log.Println(args...)
}

func ContextWithTraceID(ctx context.Context) context.Context {
	traceID := uuid.NewString()
	return context.WithValue(ctx, ctxKeyTraceID, traceID)
}

func TraceIdFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(ctxKeyTraceID).(string); ok {
		return traceID
	}
	return uuid.NewString()
}
