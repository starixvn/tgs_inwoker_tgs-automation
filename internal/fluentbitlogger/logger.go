package fluentbitlogger

import (
	"log"
	"os"
	"time"
)

type Logger struct {
	file   *os.File
	logger *log.Logger
}

// NewLogger 創建一個新的日誌記錄器，並將日誌寫入指定的文件路徑
func NewLogger(filePath string) (*Logger, error) {
	// 打開或創建日誌文件
	logFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	// 創建日誌記錄器
	logger := log.New(logFile, "", log.LstdFlags)

	return &Logger{
		file:   logFile,
		logger: logger,
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
	l.logger.Println("{\"level\":\"" + level + "\", \"message\":\"" + message + "\", \"time\":\"" + time.Now().Format(time.RFC3339) + "\"}")
}

// Close 關閉日誌文件
func (l *Logger) Close() error {
	return l.file.Close()
}
