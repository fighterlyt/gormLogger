package gormlogger

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
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

func NewLogger(logger log.Logger, slowThreshold time.Duration) logger.Interface {
	return &Logger{Logger: logger, slowThreshold: slowThreshold}
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

// 	callbacks.go replace c.processor.db.Logger.Info(context.Background(), "replacing callback `%v` from %v\n", name, utils.FileWithLineNum())
func (l Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.Info(fmt.Sprintf(msg, data...))
}

func (l Logger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.Logger.Warn(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

func (l Logger) Error(ctx context.Context, s string, i ...interface{}) {
	l.Logger.Error(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	switch {
	case err != nil:
		value := ctx.Value(IgnoreErrorKey)
		if value != nil {
			if ignoreError, ok := value.(error); ok {
				if err == ignoreError {
					return
				}
			}
		}
		l.Logger.Error(`执行错误`, zap.String(`错误`, err.Error()), zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		l.Logger.Warn(`慢查询`, zap.Duration(`阈值`, l.slowThreshold), zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))

	default:
		l.Logger.Info(`执行成功`, zap.Int64(`影响行数`, rows), zap.Duration(`耗时`, elapsed), zap.String(sqlField, sql))
	}
}
func (l Logger) gormFields(msg string, data ...interface{}) []zap.Field {
	return []zap.Field{
		zap.String(`信息`, fmt.Sprintf(msg, data...)),
	}
}

var (
	sqlField = `SQL`
	lineKey  = `gormLine`
	fileKey  = `gormFile`
)

type CtxKey string

const (
	IgnoreErrorKey = CtxKey(`ignoreError`)
)

var (
	gormPackage    = filepath.Join("gorm.io", "gorm")
	zapgormPackage = filepath.Join("fighterlyt", "gormlogger")
)

func (l *Logger) AutoSkip() {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, zapgormPackage):
		default:
			l.Logger = l.AddCallerSkip(i)
		}
	}

}
