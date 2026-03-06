package service

import (
	"Shortener/internal/models"
	"Shortener/pkg/encode"
	"context"
	"errors"
	"time"
)

type ShortenerRepositoryInterface interface {
	CreateUrl(url *models.Url) (int, error)
	UpdateUrl(url *models.Url) error
	GetUrl(shortUrl string) (string, error)
	RegisterClick(click *models.Click) error
	GetAllClicks(shortUrl string) ([]*models.Click, error)
	GetClicksAggregated(shortUrl string, groupBy string) (*models.AnalyticsResponse, error)
}

type ShortenerService struct {
	ctx  context.Context
	repo ShortenerRepositoryInterface
}

func NewShortenerService(ctx context.Context, repo ShortenerRepositoryInterface) *ShortenerService {
	return &ShortenerService{
		ctx:  ctx,
		repo: repo,
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
		return "", err
	}
	url.ShortUrl = encode.EncodeBase62(id)
	err = s.repo.UpdateUrl(url)
	if err != nil {
		return "", err
	}
	return url.ShortUrl, nil
}

func (s *ShortenerService) Redirect(shortUrl string, userAgent string) (string, error) {
	if shortUrl == "" {
		return "", errors.New("url is empty")
	}
	originalUrl, err := s.repo.GetUrl(shortUrl)
	if err != nil {
		return "", err
	}
	click := &models.Click{
		ShortUrl:  shortUrl,
		UserAgent: userAgent,
		ClickedAt: time.Now(),
	}
	_ = s.repo.RegisterClick(click)
	return originalUrl, nil
}

func (s *ShortenerService) GetAnalytics(shortUrl string, groupBy string) (*models.AnalyticsResponse, error) {
	if shortUrl == "" {
		return nil, errors.New("url is empty")
	}
	response := &models.AnalyticsResponse{
		ShortURL: shortUrl,
	}
	var err error
	switch groupBy {
	case "day", "month", "user_agent":
		response, err = s.repo.GetClicksAggregated(shortUrl, groupBy)
	default:
		clicks, err := s.repo.GetAllClicks(shortUrl)
		if err != nil {
			return nil, err
		}
		response.Clicks = clicks
		response.TotalClicks = len(clicks)
		return response, nil
	}
	if err != nil {
		return nil, err
	}

	return response, nil
}
