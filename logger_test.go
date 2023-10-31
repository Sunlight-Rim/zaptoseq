package zaptoseq

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"errors"
	"os"
	"testing"
)

func TestHookIntegration(t *testing.T) {
	hook, err := NewHook("http://localhost:5341/", "")
	if err != nil {
		t.Error(err)
	}

	log := hook.NewLoggerWith(zap.NewDevelopmentConfig(), zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	))

	log.Debug("Debug message", zap.String("level", "debug"), zap.Bool("ok", true))
	log.Info("Info message", zap.String("level", "info"), zap.Binary("binary", []byte("hello")), zap.String("original", "hello"))
	log.Warn("Warning message", zap.String("newline", "{\n    \"hello\": \"world\"\n}"))
	log.Error("Error message", zap.Error(errors.New("oh no!")))

	hook.Wait()
}
