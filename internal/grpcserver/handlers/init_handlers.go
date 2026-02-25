package grpchandlers

import (
	"github.com/IvanOplesnin/BotTradeService.git/gen/authv1"
	"github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interceptor/authinterceptor"
	"github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interceptor/loggerinterceptor"
	grpcports "github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interface"
	"google.golang.org/grpc"
)

type InitHandlerDeps struct {
	AuthUseCase   grpcports.AuthUsecase
	BotVerifier   grpcports.BotVerifier
	TokenVerifier grpcports.TokenVerifier
}

func InitHandlers(deps InitHandlerDeps) *grpc.Server {
	authHandler := NewAuthHandler(deps.AuthUseCase)
	authInterceptor := authinterceptor.NewAuthInterceptor(authinterceptor.AuthInterceptorDeps{
		BotVerifier:   deps.BotVerifier,
		TokenVerifier: deps.TokenVerifier,
	})
	loggerInterceptor := loggerinterceptor.NewLoggerInterceptor()

	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggerInterceptor.Unary(),
			authInterceptor.Unary(),
		),
	)

	authv1.RegisterAuthServiceServer(server, authHandler)
	return server
}
