// Package zaptoseq provides a hook to send logs from Zap logger to Seq (https://datalust.co/seq).
package zaptoseq

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ErrEmptyURL = errors.New("empty Seq url")
var ErrRequestCreation = errors.New("cannot create a request to Seq")

type Hook struct {
	client       *http.Client
	seqApiURL    string
	seqApiHeader http.Header
	wg           *sync.WaitGroup

	// Fallback-logger in case if Seq request fails
	fallbackLogger *zap.Logger
}

// NewHook creates a hook to Seq.
func NewHook(sequrl, token string) (*Hook, error) {
	if sequrl == "" {
		return nil, ErrEmptyURL
	}

	// Remove last URL slash.
	if sequrl[len(sequrl)-1] == '/' {
		sequrl = string(sequrl[:len(sequrl)-1])
	}

	header := make(http.Header)
	header.Set("Content-Type", "application/vnd.serilog.clef")
	if token != "" {
		header.Set("X-Seq-ApiKey", token)
	}

	return &Hook{
		client:       new(http.Client),
		seqApiURL:    fmt.Sprintf("%s/api/events/raw", sequrl),
		seqApiHeader: header,
		wg:           new(sync.WaitGroup),
	}, nil
}

// EnableFallbackLogs turns on sending errors during Seq request to the console.
func (h *Hook) EnableFallbackLogs() {
	h.fallbackLogger = zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentConfig().EncoderConfig),
		zapcore.Lock(os.Stderr),
		zap.DebugLevel,
	))
}

// DisableFallbackLogs turns off sending errors during Seq request to the console.
func (h *Hook) DisableFallbackLogs() {
	h.fallbackLogger = nil
}

// NewLogger builts a Zap-logger that send logs just to Seq.
func (h *Hook) NewLogger(zapconfig zap.Config) *zap.Logger {
	return zap.New(h.NewCore(zapconfig))
}

// NewLoggerWith builts a Zap-logger that send logs to Seq and also to other cores.
func (h *Hook) NewLoggerWith(zapconfig zap.Config, cores ...zapcore.Core) *zap.Logger {
	return zap.New(zapcore.NewTee(append(cores, h.NewCore(zapconfig))...))
}

// NewCore returns Zap core that sending logs to Seq.
func (h *Hook) NewCore(zapconfig zap.Config) zapcore.Core {
	// Seq requirement fields and values format
	zapconfig.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	zapconfig.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	zapconfig.EncoderConfig.LevelKey = "@l"
	zapconfig.EncoderConfig.TimeKey = "@t"
	zapconfig.EncoderConfig.MessageKey = "@mt"
	zapconfig.EncoderConfig.CallerKey = "caller"
	zapconfig.EncoderConfig.StacktraceKey = "trace"

	return zapcore.NewCore(zapcore.NewJSONEncoder(zapconfig.EncoderConfig), zapcore.AddSync(h), zap.DebugLevel)
}

// Write writes log to Seq. Implements the Writer interface.
func (h *Hook) Write(p []byte) (n int, err error) {
	// Since we immediately return, we need to make a copy of the payload that takes time to be sent
	req, err := http.NewRequest(http.MethodPost, h.seqApiURL, bytes.NewBuffer(append(make([]byte, 0, len(p)), p...)))
	if err != nil {
		return 0, errors.Wrap(ErrRequestCreation, err.Error())
	}
	req.Header = h.seqApiHeader

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()

		resp, err := h.client.Do(req)
		if err != nil {
			if h.fallbackLogger != nil {
				h.fallbackLogger.Error("Failed doing Seq request or reading response", zap.Error(err))
			}
			return
		}
		defer resp.Body.Close()

		// The status is supposed to be 201 (Created)
		if resp.StatusCode == 201 || h.fallbackLogger == nil {
			return
		}

		// If not, then parse a message
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			h.fallbackLogger.Error(
				"Failed reading Seq response body",
				zap.Error(errors.Wrapf(err, "status code %d", resp.StatusCode)),
			)
			return
		}

		h.fallbackLogger.Error(
			"Seq error",
			zap.String("error-message", gjson.GetBytes(content, "Error").String()),
			zap.String("raw-content", string(content)),
			zap.String("content-type", resp.Header.Get("Content-Type")),
		)
	}()

	return len(p), nil // Always success (but request might have failed)
}

// Wait until all requests are completed.
func (h *Hook) Wait() {
	h.wg.Wait()
}
