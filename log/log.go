package log

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"strings"
)

type ctxKey string

const ctxKeyRequestID ctxKey = "request_id"

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
	reqID := GetRequestID(ctx)
	prefix := fmt.Sprintf("[%s] [Request ID: %s] ", strings.ToUpper(string(severity)), reqID)
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
	reqID := GetRequestID(ctx)
	prefix := fmt.Sprintf("[%s] [RequestID: %s] ", strings.ToUpper(string(severity)), reqID)
	args := make([]any, 0, len(a)+1)
	args = append(args, prefix)
	for _, arg := range a {
		args = append(args, arg)
	}
	log.Println(args...)
}

func ContextWithRequestID(ctx context.Context) context.Context {
	requestID := uuid.NewString()
	return context.WithValue(ctx, ctxKeyRequestID, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(ctxKeyRequestID).(string); ok {
		return requestID
	}
	return uuid.NewString()
}
