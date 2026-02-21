package testdata

import (
	"log/slog"

	"go.uber.org/zap"
)

func some() {
	log := zap.Logger{}
	// Лог пишется со строчной буквы
	slog.Warn("Hello") // want "log messages must start with lowercase letter"
	slog.Debug("hello")
	log.Fatal("Hello") // want "log messages must start with lowercase letter"
	log.Info("hello")

	// Логи содержат исключительно латинские буквы
	slog.Warn("привeт") // want "log messages must only contains latin letters"
	slog.Debug("heLlo1")
	log.Fatal("привeт") // want "log messages must only contains latin letters"
	log.Info("heLlo1")

	// Никаких спецсимволов
	slog.Warn("hello!")    // want "log messages must not contains any special symbols"
	slog.Debug("hello...") // want "log messages must not contains any special symbols"
	slog.Info("hello❤️")   // want "log messages must not contains any special symbols"
	log.Warn("hello!")     // want "log messages must not contains any special symbols"
	log.Debug("hello...")  // want "log messages must not contains any special symbols"
	log.Info("hello❤️")    // want "log messages must not contains any special symbols"

	// Потенциально чувствительные данные
	token := "abracadabra"
	password := "123123"
	apiKey := "saymyname"
	slog.Warn("hello" + token)     // want "potentially sensitive data \"token\" is concatenated into log message"
	slog.Debug("hello" + password) // want "potentially sensitive data \"password\" is concatenated into log message"
	slog.Debug("hello" + apiKey)   // want "potentially sensitive data \"apiKey\" is concatenated into log message"
	log.Warn("hello" + token)      // want "potentially sensitive data \"token\" is concatenated into log message"
	log.Debug("hello" + password)  // want "potentially sensitive data \"password\" is concatenated into log message"
	log.Debug("hello" + apiKey)    // want "potentially sensitive data \"apiKey\" is concatenated into log message"

	// Ультимативно плохой лог!
	slog.Info("Очень плохо, never write logs loke this!!!!🥶" + token)         // want "potentially sensitive data \"token\" is concatenated into log message"
	log.Info("Очень плохо, never write " + password + " logs loke this!!!!🥶") // want "potentially sensitive data \"password\" is concatenated into log message"
}
