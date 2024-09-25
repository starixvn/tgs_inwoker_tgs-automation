package fluentbitlogger

import (
	"fmt"
	"log"

	"github.com/fluent/fluent-logger-golang/fluent"
)

type Logger struct {
	logger *fluent.Fluent
	tag    string
}

// NewLogger 使用 fluent-logger-golang 初始化 Fluent Bit Logger
func NewLogger(fluentBitIP string, port int, serviceName string) (*Logger, error) {
	config := fluent.Config{
		FluentPort: port,
		FluentHost: fluentBitIP,
	}

	logger, err := fluent.New(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Fluent Bit: %w", err)
	}

	return &Logger{
		logger: logger,
		tag:    serviceName,
	}, nil
}

// Info 記錄 Info 級別的日誌
func (l *Logger) Info(message string) {
	l.log("INFO", message)
}

// Error 記錄 Error 級別的日誌
func (l *Logger) Error(message string) {
	l.log("ERROR", message)
}

// Fatal 記錄 Fatal 級別的日誌
func (l *Logger) Fatal(message string) {
	l.log("FATAL", message)
}

// log 是內部方法，用於記錄日誌並附帶時間戳
func (l *Logger) log(level string, message string) {
	data := map[string]string{
		"level":   level,
		"message": message,
	}

	err := l.logger.Post(l.tag, data)
	if err != nil {
		log.Printf("failed to send log to Fluent Bit: %v", err)
	}
}

// Close 關閉與 Fluent Bit 的連接
func (l *Logger) Close() error {
	return l.logger.Close()
}
