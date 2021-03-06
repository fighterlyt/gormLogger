package gormlogger

import (
	"context"
	"testing"
	"time"

	"github.com/fighterlyt/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	localDSN     = `root:dubaihell@tcp(127.0.0.1:3306)/test1?charset=utf8mb4&parseTime=True`
	originLogger log.Logger
	err          error
	db           *gorm.DB
)

var (
	targetLogger logger.Interface
)

func TestNewLogger(t *testing.T) {
	cfg := &log.Config{
		Debug:   true,
		Service: "测试",
		Level:   zapcore.DebugLevel,
	}

	originLogger, err = cfg.Build()

	require.NoError(t, err, `构建原始日志器`)

	mysqlLogger := originLogger.Derive(`mysql`)
	mysqlLogger.Info(`before`)
	mysqlLogger = mysqlLogger.AddCallerSkip(4)
	targetLogger = NewLogger(mysqlLogger, time.Second, map[string]zapcore.Level{
		`test`: zapcore.WarnLevel,
	})

	db, err = gorm.Open(mysql.Open(localDSN), &gorm.Config{
		Logger: targetLogger,
	})

	var (
		count int64
	)

	require.NoError(t, db.WithContext(context.WithValue(todo, ModuleKey, `test`)).Table(`user_wallet`).Count(&count).Error, `COUNT`)
}
