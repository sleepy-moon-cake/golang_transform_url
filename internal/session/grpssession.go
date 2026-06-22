package session

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCJWTSession — gRPC аналог вашей мидлвари JWTSession
func GRPCJWTSession(cfg *SessionConfig) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		md, ok := metadata.FromIncomingContext(ctx)
		var tokenStr string

		if ok {
			if values := md.Get("authorization"); len(values) > 0 {
				tokenStr = values[0]
			}
		}

		if tokenStr == "" {
			newUserID := uuid.NewString()

			jwtString, err := buildJWTString(newUserID, cfg.SecretKey, cfg.ExpiresAt)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to generate session")
			}

			replyMetadata := metadata.Pairs("authorization", jwtString)
			if err := grpc.SetHeader(ctx, replyMetadata); err != nil {
				return nil, status.Error(codes.Internal, "failed to set response header")
			}

			ctx = context.WithValue(ctx, contextkeys.UserId, newUserID)
			return handler(ctx, req)
		}

		userID, err := parseJWTString(tokenStr, cfg.SecretKey)
		if err != nil {
			slog.Error("Parse gRPC JWT err", "err", err)
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		ctx = context.WithValue(ctx, contextkeys.UserId, userID)
		return handler(ctx, req)
	}
}
