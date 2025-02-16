package database

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	gormLogger "gorm.io/gorm/logger"
)

// LogrusLogger is a custom GORM logger that uses logrus
type LogrusLogger struct {
	LogLevel gormLogger.LogLevel // GORM log level
	Logger   *logrus.Entry       // Logrus logger instance
}

// New creates a new instance of LogrusLogger
func New(logLevel gormLogger.LogLevel, logrusLogger *logrus.Logger) *LogrusLogger {
	return &LogrusLogger{
		LogLevel: logLevel,
		Logger:   logrusLogger.WithField("component", "gorm"),
	}
}

// LogMode sets the logging level for GORM
func (l *LogrusLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return &LogrusLogger{
		LogLevel: level,
		Logger:   l.Logger,
	}
}

// Info logs non-critical information
func (l *LogrusLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Info {
		l.Logger.WithContext(ctx).Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warnings
func (l *LogrusLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		l.Logger.WithContext(ctx).Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs errors
func (l *LogrusLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		l.Logger.WithContext(ctx).Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL statements, execution time, and errors
func (l *LogrusLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel > gormLogger.Silent {
		elapsed := time.Since(begin)
		sql, rows := fc()

		fields := logrus.Fields{
			"duration_ms": float64(elapsed.Microseconds()) / 1000.0, // Log in milliseconds
			"rows":        rows,
			"sql":         sql,
		}

		if err != nil {
			fields["error"] = err
			l.Logger.WithContext(ctx).WithFields(fields).Error("SQL Execution failed")
		} else if elapsed > 200*time.Millisecond { // Warn for slow queries
			l.Logger.WithContext(ctx).WithFields(fields).Warn("Slow SQL query")
		} else {
			l.Logger.WithContext(ctx).WithFields(fields).Info("SQL executed")
		}
	}
}
