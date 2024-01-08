# Zap to Seq

A hook for sending [Zap](https://pkg.go.dev/go.uber.org/zap) logs to [Seq](https://datalust.co/seq).

### Documentation

Code documentation avaliable at [godoc](https://pkg.go.dev/github.com/Sunlight-Rim/zaptoseq).

### Usage

Import:
```go
import zaptoseq "github.com/Sunlight-Rim/zaptoseq"
```
\
With one core:
```go
    hook, err := zaptoseq.NewHook("http://localhost:5341", "token")
    if err != nil {
   	   panic(err)
    }

    log := hook.NewLogger(zap.NewDevelopmentConfig())

    log.Info("Hello, World!")

    hook.Wait()
```
Note: token can be an empty string.

\
With multiple cores:
```go
    // Some Zap core
    stdoutCore := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(os.Stdout),
        zapcore.DebugLevel,
    )

    hook, err := zaptoseq.NewHook("http://localhost:5341", "token")
    if err != nil {
   	   panic(err)
    }

    log := hook.NewLoggerWith(zap.NewDevelopmentConfig(), stdoutCore)

    // Will be sent to both Seq and stdout
    log.Info("Hello, World!")

    hook.Wait()
```
\
With multiple cores without DI:
```go
    // Some Zap core
    stdoutCore := zapcore.NewCore(
        zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(os.Stdout),
        zapcore.DebugLevel,
    )

    hook, err := zaptoseq.NewHook("http://localhost:5341", "token")
    if err != nil {
   	   panic(err)
    }

    // Zap logger with Seq core and stdout core
    log := zap.New(zapcore.NewTee(
        stdoutCore,
        hook.NewCore(zap.NewDevelopmentConfig()),
    ), zap.AddCaller())

    // Will be sent to both Seq and stdout
    log.Info("Hello, World!")

    hook.Wait()
```
