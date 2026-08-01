package service

import (
	"Shortener/internal/models"
	"Shortener/internal/suberrors"
	"Shortener/pkg/encode"
	"Shortener/pkg/logger"
	rediscache "Shortener/pkg/redis"
	"context"
	"errors"
	"time"

	"github.com/wb-go/wbf/redis"
	"go.uber.org/zap"
)

const cacheTTL = 24 * time.Hour

type ShortenerRepositoryInterface interface {
	CreateUrl(url *models.Url) (int, error)
	UpdateUrl(url *models.Url) error
	GetUrl(shortUrl string) (string, error)
	RegisterClick(click *models.Click) error
	GetAllClicks(shortUrl string) ([]*models.Click, error)
	GetClicksAggregated(shortUrl string, groupBy string) (*models.AnalyticsResponse, error)
}

type ShortenerService struct {
	ctx   context.Context
	repo  ShortenerRepositoryInterface
	cache *redis.Client
}

func NewShortenerService(ctx context.Context, repo ShortenerRepositoryInterface, cache *redis.Client) *ShortenerService {
	return &ShortenerService{
		ctx:   ctx,
		repo:  repo,
		cache: cache,
	}
}

func (s *ShortenerService) CreateUrl(url *models.Url) (string, error) {
	if url == nil {
		return "", errors.New("url is nil")
	}
	if url.OriginalUrl == "" {
		return "", errors.New("origin url is empty")
	}
	url.CreatedAt = time.Now()
	id, err := s.repo.CreateUrl(url)
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error("Failed to create URL in database",
			zap.String("original_url", url.OriginalUrl),
			zap.Error(err))
		return "", err
	}
	url.ShortUrl = encode.EncodeBase62(id)
	url.ID = id
	err = s.repo.UpdateUrl(url)
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error("Failed to update short URL",
			zap.Int("id", id),
			zap.String("short_url", url.ShortUrl),
			zap.Error(err))
		return "", err
	}
	logger.GetLoggerFromCtx(s.ctx).Info("Short URL created successfully",
		zap.String("short_url", url.ShortUrl),
		zap.String("original_url", url.OriginalUrl))
	return url.ShortUrl, nil
}

func (s *ShortenerService) Redirect(shortUrl string, userAgent string) (string, error) {
	if shortUrl == "" {
		return "", suberrors.ShortURLIsEmpty
	}

	originalUrl, fromCache := s.getFromCache(shortUrl)
	if !fromCache {
		var err error
		originalUrl, err = s.repo.GetUrl(shortUrl)
		if err != nil {
			logger.GetLoggerFromCtx(s.ctx).Error("Short URL not found for redirect",
				zap.String("short_url", shortUrl),
				zap.Error(err))
			return "", err
		}
		s.setCache(shortUrl, originalUrl)
	}

	click := &models.Click{
		ShortUrl:  shortUrl,
		UserAgent: userAgent,
		ClickedAt: time.Now(),
	}
	if err := s.repo.RegisterClick(click); err != nil {
		logger.GetLoggerFromCtx(s.ctx).Warn("Failed to register click",
			zap.String("short_url", shortUrl),
			zap.Error(err))
	}
	logger.GetLoggerFromCtx(s.ctx).Info("Redirect performed",
		zap.String("short_url", shortUrl),
		zap.String("original_url", originalUrl))
	return originalUrl, nil
}

func (s *ShortenerService) GetAnalytics(shortUrl string, groupBy string) (*models.AnalyticsResponse, error) {
	if shortUrl == "" {
		return nil, suberrors.ShortURLIsEmpty
	}
	response := &models.AnalyticsResponse{
		ShortURL: shortUrl,
	}
	var err error
	switch groupBy {
	case "day", "month", "user_agent":
		response, err = s.repo.GetClicksAggregated(shortUrl, groupBy)
	default:
		var clicks []*models.Click
		clicks, err = s.repo.GetAllClicks(shortUrl)
		if err != nil {
			return nil, err
		}
		response.Clicks = clicks
		response.TotalClicks = len(clicks)
		return response, nil
	}
	if err != nil {
		logger.GetLoggerFromCtx(s.ctx).Error("Failed to get clicks:", zap.Error(err))
		return nil, err
	}

	return response, nil
}

func (s *ShortenerService) getFromCache(shortUrl string) (string, bool) {
	if s.cache == nil {
		return "", false
	}
	originalUrl, err := s.cache.Get(s.ctx, rediscache.CacheKey(shortUrl))
	if err != nil {
		if !errors.Is(err, redis.NoMatches) {
			logger.GetLoggerFromCtx(s.ctx).Warn("Failed to read from cache",
				zap.String("short_url", shortUrl),
				zap.Error(err))
		}
		return "", false
	}
	return originalUrl, true
}

func (s *ShortenerService) setCache(shortUrl string, originalUrl string) {
	if s.cache == nil {
		return
	}
	if err := s.cache.SetWithExpiration(s.ctx, rediscache.CacheKey(shortUrl), originalUrl, cacheTTL); err != nil {
		logger.GetLoggerFromCtx(s.ctx).Warn("Failed to write to cache",
			zap.String("short_url", shortUrl),
			zap.Error(err))
	}
}
