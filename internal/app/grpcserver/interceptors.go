package grpcserver

import (
	"context"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/nasik90/url-shortener/internal/app/logger"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
	trustednet "github.com/nasik90/url-shortener/internal/app/trustedNet"
)

var (
	ErrUserIDMissing = errors.New("user-id is missing in metadata")
	ErrXRealIPMissed = errors.New("X-Real-IP missed")
)

// UnaryInterceptor извлекает user-id из метаданных и кладёт в контекст.
func userIDUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrUserIDMissing
	}

	userIDs := md.Get("userId")
	if len(userIDs) == 0 || userIDs[0] == "" {
		return nil, ErrUserIDMissing
	}
	userID := userIDs[0]

	// Кладём userID в контекст
	ctx = context.WithValue(ctx, middleware.UserIDContextKey{}, userID)

	// Вызываем следующий обработчик с обновлённым контекстом
	return handler(ctx, req)
}

// loggingInterceptor — unary interceptor для логирования вызовов gRPC.
func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	duration := time.Since(start)
	if err != nil {
		logger.Log.Sugar().Errorln(
			"method", info.FullMethod,
			"error", err.Error(),
			"duration", duration,
		)
	} else {
		logger.Log.Sugar().Infoln(
			"method", info.FullMethod,
			"duration", duration,
		)
	}

	return resp, err
}

// loggingInterceptor — unary interceptor для проверки доверенной сети.
func trustedNetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrXRealIPMissed
		}

		XRealIPs := md.Get("X-Real-IP")
		if len(XRealIPs) == 0 || XRealIPs[0] == "" {
			return nil, ErrXRealIPMissed
		}
		XRealIP := XRealIPs[0]

		err := trustednet.CheckForTrustedNet(trustedSubnet, XRealIP, info.FullMethod)
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}
