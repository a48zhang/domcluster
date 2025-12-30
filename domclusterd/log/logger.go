package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// Init 初始化日志
func Init(debug bool) error {
	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

// Fatal 记录致命错误日志并退出
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}


