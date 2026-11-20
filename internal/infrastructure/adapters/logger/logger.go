package logger

import (
	"context"
	"log"

	"github.com/edgarpezoa/go-template/internal/domain/ports"
)

type LoggerAdapter struct {
	logger *log.Logger
}

func NewLoggerAdapter(logger *log.Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: logger}
}

func (logger *LoggerAdapter) Log(_ context.Context, message string) {
	logger.logger.Println(message)
}

var _ ports.LoggerPort = (*LoggerAdapter)(nil)
