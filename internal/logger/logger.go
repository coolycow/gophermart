package logger

import (
	"go.uber.org/zap"
)

// Log будет доступен всему коду как синглтон.
// Никакой код навыка, кроме функции Initialize, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логгер, который не выводит никаких сообщений.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логгера с необходимым уровнем логирования.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)

	if err != nil {
		return err
	}

	// создаём новую конфигурацию логгера
	cfg := zap.NewProductionConfig()

	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логгер на основе конфигурации

	zl, err := cfg.Build()

	if err != nil {
		return err
	}

	// устанавливаем синглтон
	Log = zl

	return nil
}
