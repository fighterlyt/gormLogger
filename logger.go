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

/*NewLogger 构建日志器
参数:
*	originLogger    	log.Logger      	原始日志器，不能为nil
*	slowThreshold   	time.Duration   	慢查询阈值
返回值:
*	logger.Interface	logger.Interface	gorm日志器
*/
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

/*Info  日志
参数:
*	_	context.Context	上下文
*	s	string          内容
*	i	...interface{} 	数据
返回值:
*/
func (l Logger) Info(_ context.Context, msg string, data ...interface{}) {
	l.Logger.Info(fmt.Sprintf(msg, data...))
}

/*Warn  警告
参数:
*	_	context.Context	上下文
*	s	string          内容
*	i	...interface{} 	数据
返回值:
*/
func (l Logger) Warn(_ context.Context, s string, i ...interface{}) {
	l.Logger.Warn(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

/*Error  错误
参数:
*	_	context.Context	上下文
*	s	string          内容
*	i	...interface{} 	数据
返回值:
*/
func (l Logger) Error(_ context.Context, s string, i ...interface{}) {
	l.Logger.Error(s, zap.Any(`值`, append([]interface{}{utils.FileWithLineNum()}, i...)))
}

/*Trace 追踪
参数:
*	_    	context.Context       	上下文
*	begin	time.Time             	开始时间
*	fc   	func() (string, int64)	返回sql语句和影响行数
*	err  	error                 	执行过程中是否报错
返回值:
*/
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
