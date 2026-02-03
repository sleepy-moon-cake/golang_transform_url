package logger

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

type loggerResponseData struct {
	status int
	size   int
}

type loggerResponse struct {
	http.ResponseWriter
	loggerData *loggerResponseData
}

func (l *loggerResponse) WriteHeader(statusCode int) {
	l.loggerData.status = statusCode
	l.ResponseWriter.WriteHeader(statusCode)
}

func (l *loggerResponse) Write(body []byte) (int, error) {
	n, err := l.ResponseWriter.Write(body)

	l.loggerData.size += n

	return n, err
}

func Init(level string) {
	options := slog.HandlerOptions{
		Level: parseStringToLogLevel(level),
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &options))

	slog.SetDefault(logger)
}

func Logger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		slog.Info("Start request")

		start := time.Now()

		loggerData := loggerResponseData{status: 200}

		lr := loggerResponse{
			ResponseWriter: res,
			loggerData:     &loggerData,
		}

		h.ServeHTTP(&lr, req)

		duration := time.Since(start)

		slog.Info("Incoming request",
			slog.String("method", req.Method),
			slog.String("path", req.URL.Path),
			slog.Duration("duration", duration))

		slog.Info("Response",
			slog.Int("status", loggerData.status),
			slog.Int("size", loggerData.size))

	})
}

func parseStringToLogLevel(str string) slog.Level {
	var logLevel slog.Level
	switch str {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	default:
		logLevel = slog.LevelInfo
	}
	return logLevel
}
