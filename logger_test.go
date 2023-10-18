package zaptoseq

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"errors"
	"testing"
)

func TestHookIntegration(t *testing.T) {
	hook, err := NewHook("http://localhost:5341/", "")
	if err != nil {
		t.Error(err)
	}

	logger := hook.NewLoggerWith(zap.NewDevelopmentConfig(), zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	))

	logger.Debug("Debug message", zap.String("level", "debug"), zap.Bool("ok", true))
	logger.Info("Info message", zap.String("level", "info"), zap.Binary("binary", []byte("hello")), zap.String("original", "hello"))
	logger.Warn("Warning message", zap.String("newline", "{\n    \"hello\": \"world\"\n}"))
	logger.Error("Error message", zap.Error(errors.New("oh no!")))

	hook.Wait()
}
