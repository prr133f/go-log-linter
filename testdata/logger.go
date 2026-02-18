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
	slog.Debug("heLlo1")
	log.Fatal("привет")
	log.Info("heLlo1")

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

	// Ультимативно плохой лог!
	slog.Info("Очень плохо, never write logs loke this!!!!🥶" + token)
	log.Info("Очень плохо, never write " + password + " logs loke this!!!!🥶")
}
