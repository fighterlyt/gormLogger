package gormlogger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"newgit.cg.xxx/go-log/log"
)

var (
	localDSN     = `root:123@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True`
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
	mysqlLogger = mysqlLogger.AddCallerSkip(1)
	targetLogger = NewLogger(mysqlLogger, time.Second)

	db, err = gorm.Open(mysql.Open(localDSN), &gorm.Config{
		Logger: targetLogger,
	})

	var (
		count int64
	)

	require.NoError(t, db.Debug().Table(`dh_pay_order`).Count(&count).Error, `COUNT`)
}
