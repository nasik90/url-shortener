package grpcserver

import (
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pb "github.com/nasik90/url-shortener/internal/app/grpcapi"
	"github.com/nasik90/url-shortener/internal/app/logger"
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

	pb.RegisterShortenerServer(s.gServer, s.shortenerServer)
	if err := s.gServer.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (s *GRPCServer) StopServer() {
	s.gServer.Stop()
}
