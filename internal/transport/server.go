package transport

import (
	"Shortener/internal/models"
	"Shortener/internal/suberrors"
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
)

type ShortenerServiceInterface interface {
	CreateUrl(url *models.Url) (string, error)
	Redirect(url string) (string, error)
	GetAnalytics(url string, groupBy string) ([]models.Click, error)
}

type ShortenerServer struct {
	ctx context.Context
	cfg *config.Config
	srv ShortenerServiceInterface
}

func NewShortenerServer(ctx context.Context, srv ShortenerServiceInterface, cfg *config.Config) *ShortenerServer {
	return &ShortenerServer{
		ctx: ctx,
		cfg: cfg,
		srv: srv,
	}
}

func (s *ShortenerServer) Run() error {
	eng := ginext.New("release")
	eng.Use(ginext.Logger())

	v1 := eng.Group("/api/v1")
	v1.POST("/shorten", s.CreateURLHandler())
	v1.GET("/s/:short_url", s.RedirectHandler())
	v1.GET("/analytics/:short_url", s.GetAnalyticsHandler())

	return eng.Run(s.cfg.GetString("HOST") + ":" + s.cfg.GetString("PORT"))
}

func (s *ShortenerServer) CreateURLHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error1"})
				return
			}
		}()
		var req *models.Url
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		shortUrl, err := s.srv.CreateUrl(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"short_url": shortUrl})
	}
}

func (s *ShortenerServer) RedirectHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error1"})
				return
			}
		}()
		shortUrl := c.Param("short_url")
		url, err := s.srv.Redirect(shortUrl)
		if err != nil {
			if errors.Is(err, suberrors.URLNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, url)
	}
}

func (s *ShortenerServer) GetAnalyticsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error1"})
				return
			}
		}()
		shortUrl := c.Param("short_url")
		groupBy := c.Query("group_by")
		analytics, err := s.srv.GetAnalytics(shortUrl, groupBy)
		if err != nil {
			if errors.Is(err, suberrors.URLNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, analytics)
	}
}
