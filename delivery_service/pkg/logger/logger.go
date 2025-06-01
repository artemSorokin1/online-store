package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(env string) (*zap.Logger, error) {
	var level zapcore.Level
	switch strings.ToLower(env) {
	case "dev", "local":
		level = zapcore.DebugLevel
	case "prod":
		level = zapcore.InfoLevel
	default:
		level = zapcore.InfoLevel
	}

	err := os.MkdirAll("/var/log/app", 0755)
	if err != nil {
		panic("cannot create log directory: " + err.Error())
	}

	// Конфигурация ротации логов через lumberjack
	logFile := &lumberjack.Logger{
		Filename:   "/var/log/app/app.log",
		MaxSize:    10, // MB
		MaxBackups: 5,  // количество архивов
		MaxAge:     7,  // дней
	}

	// Конфигурация encoder'а
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:      "time",
		LevelKey:     "level",
		MessageKey:   "msg",
		CallerKey:    "caller",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	// Комбинируем sink + encoder + уровень логов
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(logFile),
		level,
	)

	// Создаем логгер
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(level))

	// Добавим service name из переменной окружения
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName != "" {
		logger = logger.With(zap.String("service", serviceName))
	}

	fmt.Println("Logger initialized")
	fileInfo, err := os.Stat("/var/log/app/app.log")
	if err != nil {
		fmt.Println("File not found:", err)
	} else {
		fmt.Println("File exists:", fileInfo.Name())
	}

	return logger, nil
}
