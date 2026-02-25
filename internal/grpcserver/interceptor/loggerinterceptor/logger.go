package loggerinterceptor

import (
	"context"
	"time"

	// <-- замени на свой logger пакет

	l "github.com/IvanOplesnin/BotTradeService.git/internal/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// LoggerInterceptor логирует каждый unary запрос.
type LoggerInterceptor struct{}

// NewLoggerInterceptor constructor (опционально)
func NewLoggerInterceptor() *LoggerInterceptor {
	return &LoggerInterceptor{}
}

func (i *LoggerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		start := time.Now()

		method := info.FullMethod

		// Если нужно — вытащим peer адрес (кто вызвал)
		// p, _ := peer.FromContext(ctx)

		// Размер запроса/ответа (примерно; актуально, если req/resp - proto messages)
		reqSize := protoSize(req)

		resp, err = handler(ctx, req)

		st := status.Convert(err)
		code := st.Code()
		// _ := st.Message()

		respSize := protoSize(resp)
		duration := time.Since(start)

		// Чтобы level был “умнее”: ошибки -> Warn/Error
		entry := l.Log.WithFields(logrus.Fields{
			"method":      method,
			"grpc_code":   code.String(),
			"status_code": int(code), // иногда удобно
			"duration":    duration,
			"req_size":    reqSize,
			"resp_size":   respSize,
		})

		if err != nil {
			// gRPC codes: NotFound/InvalidArgument обычно Warn, Internal/Unavailable — Error
			switch code {
			case codes.InvalidArgument, codes.NotFound, codes.Unauthenticated, codes.PermissionDenied, codes.AlreadyExists:
				entry.WithField("error", err.Error()).Warn("grpc request handled")
			default:
				entry.WithField("error", err.Error()).Error("grpc request handled")
			}
			return resp, err
		}

		entry.Info("grpc request handled")
		return resp, nil
	}
}

func protoSize(v any) int {
	if v == nil {
		return 0
	}
	m, ok := v.(proto.Message)
	if !ok {
		return 0
	}
	// proto.Size быстрее чем Marshal
	return proto.Size(m)
}
