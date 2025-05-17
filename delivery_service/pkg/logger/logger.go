package logger

import (
	"context"
	"go.uber.org/zap"
)

const (
	LoggerKey   = "logger"
	RequestId   = "requestID"
	ServiceName = "service"
)

type logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
}

type Logger struct {
	Logger      *zap.Logger
	ServiceName string
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.ServiceName), zap.String(RequestId, ctx.Value(RequestId).(string)))
	l.Logger.Info(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String(ServiceName, l.ServiceName), zap.String(RequestId, ctx.Value(RequestId).(string)))
	l.Logger.Error(msg, fields...)
}

func New(serviceName string) *Logger {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil
	}
	defer zapLogger.Sync()
	return &Logger{
		Logger:      zapLogger,
		ServiceName: serviceName,
	}
}

func GetLoggerFromContext(ctx context.Context) *Logger {
	return ctx.Value(LoggerKey).(*Logger)
}
