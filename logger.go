package gormlogger

import (
	"context"
	"fmt"
	"time"

	"github.com/fighterlyt/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/utils"

	"gorm.io/gorm/logger"
)

var (
	todo = context.TODO()
)

type Logger struct {
	log.Logger
	slowThreshold time.Duration // 慢查询耗时阈值
}

func NewLogger(originLogger log.Logger, slowThreshold time.Duration) logger.Interface {
	return &Logger{Logger: originLogger, slowThreshold: slowThreshold}
}

func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	var targetLevel zapcore.Level

	switch level {
	case logger.Info:
		targetLevel = zapcore.InfoLevel
	case logger.Warn:
		targetLevel = zapcore.WarnLevel
	case logger.Error:
		targetLevel = zapcore.ErrorLevel
	case logger.Silent:
		targetLevel = zapcore.PanicLevel
	}

	l.Logger = l.Logger.SetLevel(targetLevel)

	return l
}

func (l Logger) Info(_ context.Context, msg string, data ...interface{}) {
	l.Logger.Info(fmt.Sprintf(msg, data...))
}

func (l Logger) Warn(_ context.Context, s string, i ...interface{}) {
	l.Logger.Warn(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

func (l Logger) Error(_ context.Context, s string, i ...interface{}) {
	l.Logger.Error(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

func (l Logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil:
		l.Logger.Error(`执行错误`, zap.String(`错误`, err.Error()), zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		l.Logger.Warn(`慢查询`, zap.Duration(`阈值`, l.slowThreshold), zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))

	default:
		l.Logger.Info(`执行成功`, zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
	}
}

var (
	sqlField = `SQL`
)
