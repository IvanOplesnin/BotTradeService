package authinterceptor

import (
	"context"

	"github.com/IvanOplesnin/BotTradeService.git/internal/app/models"
	modelerrors "github.com/IvanOplesnin/BotTradeService.git/internal/domain/errors"
	"github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/authctx"
	"github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/grpcutil"
	grpcports "github.com/IvanOplesnin/BotTradeService.git/internal/grpcserver/interface"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type AuthInterceptor struct {
	svcBotVerifier   grpcports.BotVerifier
	svcTokenVerifier grpcports.TokenVerifier

	PublicMethods map[string]struct{}
	BotMethods    map[string]struct{}
}

type AuthInterceptorDeps struct {
	BotVerifier   grpcports.BotVerifier
	TokenVerifier grpcports.TokenVerifier
}

func NewAuthInterceptor(deps AuthInterceptorDeps) *AuthInterceptor {
	return &AuthInterceptor{
		svcBotVerifier:   deps.BotVerifier,
		svcTokenVerifier: deps.TokenVerifier,
		PublicMethods: map[string]struct{}{
			"/bottrade.auth.v1.AuthService/Register": {},
			"/bottrade.auth.v1.AuthService/Login":    {},
		},
		BotMethods: map[string]struct{}{
			"/bottrade.auth.v1.AuthService/LinkTelegram":  {},
			"/bottrade.auth.v1.AuthService/TelegramAuth": {},
		},
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {

		// 1) public methods: no auth
		if _, ok := i.PublicMethods[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		// 2) bot methods: require bot signature metadata
		if _, ok := i.BotMethods[info.FullMethod]; ok {
			md, _ := metadata.FromIncomingContext(ctx)
			meta, err := extractBotMeta(md)
			if err != nil {
				return nil, err
			}

			reqBytes, err := marshalReqBytes(req)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to marshal request")
			}

			if err := i.svcBotVerifier.ValidateBotSignature(ctx, meta, info.FullMethod, reqBytes); err != nil {
				return nil, mapSvcErr(err)
			}

			// bot methods обычно не кладут user_id, потому что user_id определяется позже (по tg_id).
			return handler(ctx, req)
		}

		// 3) jwt required
		md, _ := metadata.FromIncomingContext(ctx)
		authz := grpcutil.GetMDString(md, "authorization")
		token, ok := grpcutil.ParseBearer(authz)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing bearer token")
		}

		userID, err := i.svcTokenVerifier.ValidateAccessToken(ctx, token)
		if err != nil {
			return nil, mapSvcErr(err)
		}

		ctx = authctx.WithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

func extractBotMeta(md metadata.MD) (models.BotMeta, error) {
	botID := grpcutil.GetMDString(md, "x-bot-id")
	tsStr := grpcutil.GetMDString(md, "x-ts")
	nonce := grpcutil.GetMDString(md, "x-nonce")
	sig := grpcutil.GetMDString(md, "x-signature")

	if botID == "" || tsStr == "" || nonce == "" || sig == "" {
		return models.BotMeta{}, status.Error(codes.Unauthenticated, "missing bot signature headers")
	}

	// Timestamp парсим максимально просто (ожидаем unix seconds)
	// Лучше в сервисе ещё проверить окно времени.
	var ts int64
	for _, ch := range tsStr {
		if ch < '0' || ch > '9' {
			return models.BotMeta{}, status.Error(codes.Unauthenticated, "bad x-ts")
		}
		ts = ts*10 + int64(ch-'0')
	}

	return models.BotMeta{
		BotID:     botID,
		Timestamp: ts,
		Nonce:     nonce,
		Signature: sig,
	}, nil
}

func marshalReqBytes(req any) ([]byte, error) {
	pm, ok := req.(proto.Message)
	if !ok {
		// fallback: deterministic hash от типа+текущего времени не нужен
		// лучше считать, что все req — proto.Message
		return nil, status.Error(codes.Internal, "request is not proto message")
	}
	// proto.Marshal для подписи достаточно; можно сделать deterministic, если нужен строгий порядок
	return proto.Marshal(pm)
}

// mapSvcErr — общий маппинг сервисных ошибок в gRPC codes
func mapSvcErr(err error) error {
	switch err {
	case modelerrors.ErrUnauthorized, modelerrors.ErrBadBotSignature:
		return status.Error(codes.Unauthenticated, err.Error())
	case modelerrors.ErrForbidden:
		return status.Error(codes.PermissionDenied, err.Error())
	case modelerrors.ErrReplay:
		return status.Error(codes.Unauthenticated, "replay detected")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
