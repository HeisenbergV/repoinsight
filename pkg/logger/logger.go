package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *logrus.Logger

type Config struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	Output     string `yaml:"output"`
	FileConfig struct {
		Path       string `yaml:"path"`
		Name       string `yaml:"name"`
		MaxSize    int    `yaml:"max_size"`
		MaxBackups int    `yaml:"max_backups"`
		MaxAge     int    `yaml:"max_age"`
		Compress   bool   `yaml:"compress"`
	} `yaml:"file"`
}

func Init(config Config) error {
	log = logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return fmt.Errorf("parse log level error: %v", err)
	}
	log.SetLevel(level)

	// 设置日志格式
	if config.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置日志输出
	if config.Output == "file" {
		// 确保日志目录存在
		if err := os.MkdirAll(config.FileConfig.Path, 0755); err != nil {
			return fmt.Errorf("create log directory error: %v", err)
		}

		// 配置日志文件
		logFile := &lumberjack.Logger{
			Filename:   filepath.Join(config.FileConfig.Path, config.FileConfig.Name),
			MaxSize:    config.FileConfig.MaxSize,    // MB
			MaxBackups: config.FileConfig.MaxBackups, // 保留的旧文件最大个数
			MaxAge:     config.FileConfig.MaxAge,     // 保留的最大天数
			Compress:   config.FileConfig.Compress,   // 是否压缩
		}

		// 同时输出到文件和控制台
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	} else {
		// 只输出到控制台
		log.SetOutput(os.Stdout)
	}

	return nil
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Warn(args ...interface{}) {
	log.Warn(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Fatal(args ...interface{}) {
	log.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}
