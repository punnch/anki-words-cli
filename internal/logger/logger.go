package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/punnch/ankiwords/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger

	file *os.File
}

func NewLogger(cfg config.Config) (*Logger, error) {
	zapLvl := zap.NewAtomicLevel()
	if err := zapLvl.UnmarshalText([]byte(cfg.LoggerLevel)); err != nil {
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	zapConfig := zap.NewDevelopmentEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000")

	// ConsoleEncoder — human format
	zapEncoder := zapcore.NewConsoleEncoder(zapConfig)

	cores := []zapcore.Core{
		zapcore.NewCore(zapEncoder, zapcore.AddSync(os.Stdout), zapLvl),
	}

	var logFile *os.File
	if err := os.MkdirAll(cfg.LoggerFolder, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "file logging disabled:", fmt.Errorf("mkdir log folder: %w", err))
	} else {
		timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
		logFilePath := filepath.Join(
			cfg.LoggerFolder,
			fmt.Sprintf("%s.log", timestamp),
		)

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "file logging disabled:", fmt.Errorf("open log file: %w", err))
		} else {
			logFile = file
			cores = append(cores, zapcore.NewCore(zapEncoder, zapcore.AddSync(logFile), zapLvl))
		}
	}

	// zap.AddCaller() - appends to every message file_name and line number
	zapLogger := zap.New(zapcore.NewTee(cores...), zap.AddCaller())

	return &Logger{
		Logger: zapLogger,
		file:   logFile,
	}, nil
}

func (l *Logger) Close() {
	if l.file == nil {
		return
	}

	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close application logger:", err)
	}
}
