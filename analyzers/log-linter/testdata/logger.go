package testdata

import (
	"log/slog"

	"go.uber.org/zap"
)

func some() {
	log := zap.Logger{}
	// Лог пишется со строчной буквы
	slog.Warn("Hello")
	slog.Debug("hello")
	log.Fatal("Hello")
	log.Info("hello")

	// Логи содержат исключительно латинские буквы
	slog.Warn("привет")
	slog.Debug("hello")
	log.Fatal("привет")
	log.Info("hello")

	// Никаких спецсимволов
	slog.Warn("hello!")
	slog.Debug("hello...")
	slog.Info("hello❤️")
	log.Warn("hello!")
	log.Debug("hello...")
	log.Info("hello❤️")

	// Потенциально чувствительные данные
	token := "abracadabra"
	password := "123123"
	apiKey := "saymyname"
	slog.Warn("hello" + token)
	slog.Debug("hello" + password)
	slog.Debug("hello" + apiKey)
	log.Warn("hello" + token)
	log.Debug("hello" + password)
	log.Debug("hello" + apiKey)
}
