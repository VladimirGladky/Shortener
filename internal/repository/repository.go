package repository

import (
	"Shortener/internal/models"
	"Shortener/internal/suberrors"
	"context"
	"database/sql"
	"errors"

	"github.com/wb-go/wbf/dbpg"
)

type ShortenerRepository struct {
	ctx context.Context
	db  *dbpg.DB
}

func NewShortenerRepository(ctx context.Context, db *dbpg.DB) *ShortenerRepository {
	return &ShortenerRepository{
		ctx: ctx,
		db:  db,
	}
}

func (s *ShortenerRepository) CreateUrl(url *models.Url) (int, error) {
	query := `INSERT INTO urls (original_url)
			  VALUES ($1) 
			  RETURNING id
    `
	var id int
	err := s.db.QueryRowContext(s.ctx, query, url.OriginalUrl).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *ShortenerRepository) GetUrl(shortUrl string) (string, error) {
	query := `SELECT original_url FROM urls WHERE short_url = $1`
	var originalUrl string
	err := s.db.QueryRowContext(s.ctx, query, shortUrl).Scan(&originalUrl)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", suberrors.URLNotFound
		}
		return "", err
	}
	return originalUrl, nil
}

func (s *ShortenerRepository) UpdateUrl(url *models.Url) error {
	query := `UPDATE urls SET short_url = $1 WHERE id = $2`
	_, err := s.db.ExecContext(s.ctx, query, url.ShortUrl, url.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *ShortenerRepository) RegisterClick(click *models.Click) error {
	query := `INSERT INTO clicks (short_url, user_agent) VALUES ($1, $2)`
	_, err := s.db.ExecContext(s.ctx, query, click.ShortUrl, click.UserAgent)
	if err != nil {
		return err
	}
	return nil
}

func (s *ShortenerRepository) GetAllClicks(shortUrl string) ([]*models.Click, error) {
	query := `SELECT id,short_url,clicked_at,user_agent FROM clicks WHERE short_url = $1 ORDER BY clicked_at DESC`
	var clicks []*models.Click
	rows, err := s.db.QueryContext(s.ctx, query, shortUrl)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		cl := &models.Click{}
		err = rows.Scan(&cl.ID, &cl.ShortUrl, &cl.ClickedAt, &cl.UserAgent)
		if err != nil {
			return nil, err
		}
		clicks = append(clicks, cl)
	}
	return clicks, nil
}

func (s *ShortenerRepository) GetClicksAggregated(shortUrl string, groupBy string) (*models.AnalyticsResponse, error) {
	response := &models.AnalyticsResponse{
		ShortURL: shortUrl,
	}

	var query string

	switch groupBy {
	case "day":
		query = `
              SELECT
                  DATE(clicked_at) as period,
                  COUNT(*) as clicks
              FROM clicks
              WHERE short_url = $1
              GROUP BY DATE(clicked_at)
              ORDER BY period DESC
          `

		rows, err := s.db.QueryContext(s.ctx, query, shortUrl)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var stats []models.DayStats
		totalClicks := 0

		for rows.Next() {
			var s models.DayStats
			if err := rows.Scan(&s.Date, &s.Clicks); err != nil {
				return nil, err
			}
			stats = append(stats, s)
			totalClicks += s.Clicks
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}

		response.ByDay = stats
		response.TotalClicks = totalClicks

	case "month":
		query = `
              SELECT
                  TO_CHAR(clicked_at, 'YYYY-MM') as period,
                  COUNT(*) as clicks
              FROM clicks
              WHERE short_url = $1
              GROUP BY TO_CHAR(clicked_at, 'YYYY-MM')
              ORDER BY period DESC
          `

		rows, err := s.db.QueryContext(s.ctx, query, shortUrl)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var stats []models.MonthStats
		totalClicks := 0

		for rows.Next() {
			var s models.MonthStats
			if err := rows.Scan(&s.Month, &s.Clicks); err != nil {
				return nil, err
			}
			stats = append(stats, s)
			totalClicks += s.Clicks
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}

		response.ByMonth = stats
		response.TotalClicks = totalClicks

	case "user_agent":
		query = `
              SELECT
                  user_agent,
                  COUNT(*) as clicks
              FROM clicks
              WHERE short_url = $1
              GROUP BY user_agent
              ORDER BY clicks DESC
          `

		rows, err := s.db.QueryContext(s.ctx, query, shortUrl)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var stats []models.UserAgentStats
		totalClicks := 0

		for rows.Next() {
			var s models.UserAgentStats
			if err := rows.Scan(&s.UserAgent, &s.Clicks); err != nil {
				return nil, err
			}
			stats = append(stats, s)
			totalClicks += s.Clicks
		}

		if err = rows.Err(); err != nil {
			return nil, err
		}

		response.ByUserAgent = stats
		response.TotalClicks = totalClicks

	default:
		return nil, errors.New("invalid groupBy parameter")
	}

	return response, nil
}
