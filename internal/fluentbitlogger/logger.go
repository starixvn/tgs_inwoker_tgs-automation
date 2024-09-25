package fluentbitlogger

import (
	"fmt"
	"log"
	"net"
	"time"
)

type Logger struct {
	conn   net.Conn
	logger *log.Logger
}

func NewLogger(fluentBitIP string, port int) (*Logger, error) {
	address := fmt.Sprintf("%s:%d", fluentBitIP, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	logger := log.New(conn, "", log.LstdFlags)

	return &Logger{
		conn:   conn,
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
	return l.conn.Close()
}
