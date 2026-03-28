package logger

import (
	domain "go-order-api/internal/domain/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	logger *zap.Logger
}

func NewZapLogger() *ZapLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.OutputPaths = []string{"stdout"}

	log, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &ZapLogger{
		logger: log,
	}
}

func (l *ZapLogger) toZapFields(fields []domain.Field) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}

func (l *ZapLogger) Info(msg string, fields ...domain.Field) {
	l.logger.Info(msg, l.toZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, err error, fields ...domain.Field) {
	if err != nil {
		fields = append(fields, domain.Any("error", err.Error()))
	}
	l.logger.Error(msg, l.toZapFields(fields)...)
}

func (l *ZapLogger) Debug(msg string, fields ...domain.Field) {
	l.logger.Debug(msg, l.toZapFields(fields)...)
}

func (l *ZapLogger) Warning(msg string, fields ...domain.Field) {
	l.logger.Warn(msg, l.toZapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, err error, fields ...domain.Field) {
	if err != nil {
		fields = append(fields, domain.Any("error", err.Error()))
	}
	l.logger.Fatal(msg, l.toZapFields(fields)...)
}

func (l *ZapLogger) Close() {
	_ = l.logger.Sync()
}
