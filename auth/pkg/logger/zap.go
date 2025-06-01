package logger

import (
	"go.uber.org/zap"
	"sync"
)

var (
	log  *zap.Logger
	once sync.Once
)

func InitLogger() {
	var err error
	once.Do(func() {
		log, err = zap.NewProduction()
	})

	if err != nil {
		panic(err)
	}
}

func GetLogger() *zap.Logger {
	if log == nil {
		InitLogger()
	}
	return log
}
