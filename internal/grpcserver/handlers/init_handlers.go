package handlers

import (
	"github.com/IvanOplesnin/BotTradeService.git/gen/authv1"
	"github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interceptor/authinterceptor"
	grpcports "github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interface"
	"google.golang.org/grpc"
)

func InitAuthHandlers(svcUseCase grpcports.AuthUsecase, authInterceptorDeps authinterceptor.AuthInterceptorDeps) *grpc.Server {
	authHandler := NewAuthHandler(svcUseCase)
	authInterceptor := authinterceptor.NewAuthInterceptor(authInterceptorDeps)

	server := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor.Unary()),
	)

	authv1.RegisterAuthServiceServer(server, authHandler)

	return server
}
