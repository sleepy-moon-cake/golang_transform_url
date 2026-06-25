package grpc

import (
	"context"

	pb "github.com/sleepy-moon-cake/golang_transform_url/internal/app/grpc/proto"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type URLService interface {
	CreateShortURL(ctx context.Context, str string) (string, error)
	GetURLByCode(ctx context.Context, code string) (string, error)
	GetUserShortUrls(ctx context.Context) ([]model.ShortenURLRecord, error)
}

type ShortenerGRPCServer struct {
	pb.UnimplementedShortenerServiceServer
	service URLService
}

func NewShortenerGRPCServer(service URLService) *ShortenerGRPCServer {
	return &ShortenerGRPCServer{
		service: service,
	}
}

func (s *ShortenerGRPCServer) ShortenURL(ctx context.Context, req *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "URL cannot be empty")
	}

	shortURL, err := s.service.CreateShortURL(ctx, req.GetUrl())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create short URL: %v", err)
	}

	return &pb.URLShortenResponse{Result: shortURL}, nil
}

func (s *ShortenerGRPCServer) ExpandURL(ctx context.Context, req *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID cannot be empty")
	}

	origURL, err := s.service.GetURLByCode(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get URL by code: %v", err)
	}

	return &pb.URLExpandResponse{Result: origURL}, nil
}

func (s *ShortenerGRPCServer) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, ok := ctx.Value(contextkeys.UserId).(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	records, err := s.service.GetUserShortUrls(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch user URLs: %v", err)
	}

	pbUrls := make([]*pb.URLData, 0, len(records))
	for _, r := range records {
		pbUrls = append(pbUrls, &pb.URLData{
			ShortUrl:    r.ShortURL,
			OriginalUrl: r.OriginalURL,
		})
	}

	return &pb.UserURLsResponse{Url: pbUrls}, nil
}
