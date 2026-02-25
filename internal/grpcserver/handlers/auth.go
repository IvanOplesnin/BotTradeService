package grpchandlers

import (
	"context"
	"net/mail"
	"strings"
	"time"

	"github.com/IvanOplesnin/BotTradeService.git/gen/authv1"
	modelerrors "github.com/IvanOplesnin/BotTradeService.git/internal/domain/errors"
	"github.com/IvanOplesnin/BotTradeService.git/internal/domain/models"
	"github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/authctx"
	grpcports "github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interface"
	"github.com/IvanOplesnin/BotTradeService.git/internal/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authv1.UnimplementedAuthServiceServer
	svc             grpcports.AuthUsecase
	codeTgTtlMinute int64
}

func NewAuthHandler(svc grpcports.AuthUsecase) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	email := strings.TrimSpace(req.GetEmail())
	pass := req.GetPassword()

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(pass); err != nil {
		return nil, err
	}

	toks, err := h.svc.Register(ctx, email, pass)
	if err != nil {
		logger.Log.Errorf("register error")
		return nil, mapAuthErr(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  toks.AccessToken,
		ExpiresInSec: toks.ExpiresInSec,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	email := strings.TrimSpace(req.GetEmail())
	pass := req.GetPassword()

	if err := validateEmail(email); err != nil {
		return nil, err
	}
	if err := validatePassword(pass); err != nil { // можно не проверять строгую длину на логине
		return nil, err
	}

	toks, err := h.svc.Login(ctx, email, pass)
	if err != nil {
		return nil, mapAuthErr(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  toks.AccessToken,
		ExpiresInSec: toks.ExpiresInSec,
	}, nil
}

func (h *AuthHandler) CreateTelegramLinkCode(ctx context.Context, _ *authv1.CreateTelegramLinkCodeRequest) (*authv1.CreateTelegramLinkCodeResponse, error) {
	userID, ok := authctx.UserID(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	ttl := time.Duration(h.codeTgTtlMinute) * time.Minute

	code, expSec, err := h.svc.CreateTelegramLinkCode(ctx, userID, ttl)
	if err != nil {
		return nil, mapAuthErr(err)
	}

	return &authv1.CreateTelegramLinkCodeResponse{
		Code:         code,
		ExpiresInSec: expSec,
	}, nil
}

func (h *AuthHandler) LinkTelegram(ctx context.Context, req *authv1.LinkTelegramRequest) (*authv1.LinkTelegramResponse, error) {
	code := strings.TrimSpace(req.GetCode())
	if code == "" {
		return nil, status.Error(codes.InvalidArgument, "code is required")
	}
	if req.GetTelegramUserId() <= 0 || req.GetChatId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "telegram_user_id/chat_id must be positive")
	}

	tg := models.TelegramProfile{
		TelegramUserID: req.GetTelegramUserId(),
		ChatID:         req.GetChatId(),
		Username:       strings.TrimSpace(req.GetUsername()),
		FirstName:      strings.TrimSpace(req.GetFirstName()),
		LastName:       strings.TrimSpace(req.GetLastName()),
	}

	if err := h.svc.LinkTelegram(ctx, code, tg); err != nil {
		return nil, mapAuthErr(err)
	}

	return &authv1.LinkTelegramResponse{Ok: true}, nil
}

func (h *AuthHandler) TelegramAuth(
	ctx context.Context,
	req *authv1.TelegramLoginRequest,
) (*authv1.AuthResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	tgUserID := req.GetTelegramUserId()
	chatID := req.GetChatId()

	if tgUserID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "telegram_user_id must be positive")
	}
	if chatID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "chat_id must be positive")
	}

	tg := models.TelegramProfile{
		TelegramUserID: tgUserID,
		ChatID:         chatID,
		Username:       strings.TrimSpace(req.GetUsername()),
		FirstName:      strings.TrimSpace(req.GetFirstName()),
		LastName:       strings.TrimSpace(req.GetLastName()),
	}

	// сервис делает "login-or-register" (если tg еще не привязан — создает юзера и привязку)
	toks, err := h.svc.TelegramAuth(ctx, tg)
	if err != nil {
		return nil, mapAuthErr(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  toks.AccessToken,
		ExpiresInSec: toks.ExpiresInSec,
	}, nil
}

// ----- Validation helpers -----

func validateEmail(email string) error {
	if email == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	_, err := mail.ParseAddress(email)
	if err != nil {
		return status.Error(codes.InvalidArgument, "email is invalid")
	}
	return nil
}

func validatePassword(p string) error {
	// MVP-ограничения
	if len(p) < 8 {
		return status.Error(codes.InvalidArgument, "password must be at least 8 characters")
	}
	if len(p) > 128 {
		return status.Error(codes.InvalidArgument, "password too long")
	}
	return nil
}

// ----- Error mapping -----

func mapAuthErr(err error) error {
	switch err {
	case modelerrors.ErrInvalidCredentials:
		return status.Error(codes.Unauthenticated, "invalid credentials")
	case modelerrors.ErrEmailTaken:
		return status.Error(codes.AlreadyExists, "email already taken")

	case modelerrors.ErrLinkCodeInvalid:
		return status.Error(codes.NotFound, "link code not found")
	case modelerrors.ErrLinkCodeExpired:
		return status.Error(codes.FailedPrecondition, "link code expired")
	case modelerrors.ErrLinkCodeUsed:
		return status.Error(codes.FailedPrecondition, "link code already used")
	case modelerrors.ErrTelegramAlreadyLinked:
		return status.Error(codes.AlreadyExists, "telegram already linked")

	case modelerrors.ErrUnauthorized, modelerrors.ErrBadBotSignature, modelerrors.ErrReplay:
		return status.Error(codes.Unauthenticated, err.Error())
	case modelerrors.ErrForbidden:
		return status.Error(codes.PermissionDenied, err.Error())

	default:
		return status.Error(codes.Internal, "internal error")
	}
}
