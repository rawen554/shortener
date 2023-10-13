package handlers

import (
	"context"

	pb "github.com/rawen554/shortener/internal/handlers/proto"
	"github.com/rawen554/shortener/internal/logic"
	"github.com/rawen554/shortener/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCService struct {
	pb.UnimplementedShortenerServer
	logger    *zap.SugaredLogger
	coreLogic *logic.CoreLogic
}

func NewService(logger *zap.SugaredLogger, coreLogic *logic.CoreLogic) *GRPCService {
	return &GRPCService{logger: logger, coreLogic: coreLogic}
}

func (gh *GRPCService) CreateShortURL(
	ctx context.Context,
	req *pb.CreateShortURLRequest,
) (*pb.CreateShortURLResponse, error) {
	url, err := gh.coreLogic.ShortenURL(ctx, req.GetUserId(), req.GetUrl())
	if err != nil {
		gh.logger.Error("shortenURL service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.CreateShortURLResponse{Result: url}, nil
}

func (gh *GRPCService) BatchCreateShortURL(
	ctx context.Context,
	req *pb.BatchCreateShortURLRequest,
) (*pb.BatchCreateShortURLResponse, error) {
	items := []models.URLBatchReq{}
	for _, item := range req.GetRecords() {
		items = append(items, models.URLBatchReq{OriginalURL: item.OriginalUrl, CorrelationID: item.CorrelationId})
	}
	res, err := gh.coreLogic.ShortenBatch(ctx, req.GetUserId(), items)
	if err != nil {
		gh.logger.Error("shortenBatch service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	result := []*pb.BatchCreateShortURLResponseData{}
	for _, item := range res {
		result = append(
			result,
			&pb.BatchCreateShortURLResponseData{
				ShortUrl:      item.ShortURL,
				CorrelationId: item.CorrelationID,
			})
	}

	return &pb.BatchCreateShortURLResponse{Records: result}, nil
}

func (gh *GRPCService) GetByShort(
	ctx context.Context,
	req *pb.GetOriginalURLRequest,
) (*pb.GetOriginalURLResponse, error) {
	originalURL, err := gh.coreLogic.GetOriginalURL(ctx, req.GetUrl())
	if err != nil {
		gh.logger.Error("redirectToOriginal service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.GetOriginalURLResponse{OriginalUrl: originalURL}, nil
}

func (gh *GRPCService) GetUserURLs(
	ctx context.Context,
	req *pb.GetUserURLsRequest,
) (*pb.GetUserURLsResponse, error) {
	urls, err := gh.coreLogic.GetUserRecords(ctx, req.GetUserId())
	if err != nil {
		gh.logger.Error("getUserRecords service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	result := []*pb.ShortenData{}
	for _, item := range urls {
		result = append(result, &pb.ShortenData{ShortUrl: item.ShortURL, OriginalUrl: item.OriginalURL})
	}
	return &pb.GetUserURLsResponse{Records: result}, nil
}

func (gh *GRPCService) DeleteUserURLsBatch(
	ctx context.Context,
	req *pb.DeleteUserURLsBatchRequest,
) (*pb.DeleteUserURLsBatchResponse, error) {
	if err := gh.coreLogic.DeleteUserRecords(ctx, req.GetUserId(), req.GetUrls()); err != nil {
		gh.logger.Error("DeleteUserURLsBatch service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.DeleteUserURLsBatchResponse{}, nil
}

func (gh *GRPCService) GetStats(
	ctx context.Context,
	req *pb.ServiceStatsRequest,
) (*pb.ServiceStatsResponse, error) {
	stats, err := gh.coreLogic.GetStats(ctx)
	if err != nil {
		gh.logger.Error("getStats service err", zap.Error(err))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.ServiceStatsResponse{Urls: int64(stats.URLs), Users: int64(stats.Users)}, nil
}
