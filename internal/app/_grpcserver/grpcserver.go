package grpcserver

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pb "github.com/nasik90/url-shortener/internal/app/grpcapi"
	"github.com/nasik90/url-shortener/internal/app/logger"
	middleware "github.com/nasik90/url-shortener/internal/app/middlewares"
)

var (
	ErrUserIDMissing = errors.New("user-id is missing in metadata")
	ErrXRealIPMissed = errors.New("X-Real-IP missed")
)

type GRPCServer struct {
	gServer         *grpc.Server
	shortenerServer *pb.ShortenerServerStruct
	serverAddress   string
}

func NewGRPCServer(shortenerServer *pb.ShortenerServerStruct, serverAddress string, trustedSubnet string) *GRPCServer {
	s := &GRPCServer{}
	s.gServer = grpc.NewServer(grpc.ChainUnaryInterceptor(loggingInterceptor, userIDUnaryInterceptor, trustedNetInterceptor(trustedSubnet)))
	s.shortenerServer = shortenerServer
	s.serverAddress = serverAddress
	return s
}

func (s *GRPCServer) RunServer() error {

	logger.Log.Info("Running grpc server", zap.String("address", s.serverAddress))

	listen, err := net.Listen("tcp", s.serverAddress)
	if err != nil {
		return err
	}

	//grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(loggingInterceptor, userIDUnaryInterceptor))

	pb.RegisterShortenerServer(s.gServer, s.shortenerServer)
	if err := s.gServer.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (s *GRPCServer) StopServer() {
	s.gServer.Stop()
}

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

var methodsToChectTrustedNet = [1]string{"GetURLsStats"}

// loggingInterceptor — unary interceptor для логирования вызовов gRPC.
func trustedNetInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		needToCheck := false
		for _, methodToCheck := range methodsToChectTrustedNet {
			if strings.Contains(info.FullMethod, methodToCheck) {
				needToCheck = true
			}
		}

		if !needToCheck {
			return handler(ctx, req)
		}

		if trustedSubnet == "" {
			return nil, errors.New("trusted subnet is empty, access forbidden")
		}

		_, trustedNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, err
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrXRealIPMissed
		}

		XRealIPs := md.Get("X-Real-IP")
		if len(XRealIPs) == 0 || XRealIPs[0] == "" {
			return nil, ErrXRealIPMissed
		}
		XRealIP := XRealIPs[0]

		ip := net.ParseIP(XRealIP)
		if ip == nil {
			return nil, errors.New("forbidden - invalid IP")
		}

		if !trustedNet.Contains(ip) {
			return nil, errors.New("forbidden - IP not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
