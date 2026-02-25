package app

import (
	"net"

	"github.com/IvanOplesnin/BotTradeService.git/internal/config"
	grpchandlers "github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/handlers"
	"github.com/IvanOplesnin/BotTradeService.git/internal/logger"
	"github.com/IvanOplesnin/BotTradeService.git/internal/repository/psql"
	"github.com/IvanOplesnin/BotTradeService.git/internal/service/hasher/argon2hash"
	"github.com/IvanOplesnin/BotTradeService.git/internal/service/svcauth"
	"github.com/IvanOplesnin/BotTradeService.git/internal/service/token"
	"google.golang.org/grpc"
)

type App struct {
	cfg *config.Config

	grpcServer *grpc.Server
	close      func()
}

func InitApp(configPath string) (*App, error) {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}

	if err := logger.SetupLogger(&cfg.Logger); err != nil {
		return nil, err
	}

	hasherPass, err := argon2hash.New(cfg.Security.PasswordHash)
	if err != nil {
		logger.Log.Errorf("no init hasher: %s", err.Error())
		return nil, err
	}
	tokener, err := token.NewTokener(cfg.Security.Tokener)
	if err != nil {
		logger.Log.Errorf("no init tokener: %s", err.Error())
		return nil, err
	}
	pool, err := psql.Connect(cfg.App.Dsn)
	if err != nil {
		logger.Log.Errorf("no init repo: %s", err.Error())
		return nil, err
	}

	repo := psql.NewPsqlRepo(pool)

	authService := svcauth.New(
		svcauth.AuthUsecaseDeps{
			Hasher:  hasherPass,
			Tokener: tokener,
			Repo:    repo,
		},
	)

	server := grpchandlers.InitHandlers(
		grpchandlers.InitHandlerDeps{
			TokenVerifier: authService,
			AuthUseCase:   authService,
			BotVerifier:   authService,
		},
	)

	return &App{
		cfg:        cfg,
		grpcServer: server,
		close: func() {
			pool.Close()
		},
	}, nil
}

func (a *App) Run() error {
	lis, err := net.Listen("tcp", a.cfg.App.Address)
	if err != nil {
		logger.Log.Errorf("app.Run error: %s", err)
		return err
	}
	defer a.close()
	if err := a.grpcServer.Serve(lis); err != nil {
		logger.Log.Errorf("app.Run error: %s", err)
		return err
	}
	return nil
}

func (a *App) Close() {
	if a.close != nil {
		a.close()
	}
}

func (a *App) GracefulStop() {
	a.grpcServer.GracefulStop()
}

func (a *App) Stop() {
	a.grpcServer.Stop()
}
