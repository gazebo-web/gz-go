package middleware

import (
	"context"
	"fmt"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
)

// LoggerGRPC adapts zap logger to interceptor logger.
// Code copied from: 
//   https://github.com/grpc-ecosystem/go-grpc-middleware/blob/a18e1e2bacb23afca0f52b228f6b4efbb5f57822/interceptors/logging/examples/zap/example_test.go#L17
func LoggerGRPC(l *zap.Logger) grpc_logging.Logger {
	return grpc_logging.LoggerFunc(func(ctx context.Context, lvl grpc_logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case grpc_logging.LevelDebug:
			logger.Debug(msg)
		case grpc_logging.LevelInfo:
			logger.Info(msg)
		case grpc_logging.LevelWarn:
			logger.Warn(msg)
		case grpc_logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
