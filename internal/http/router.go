package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devang/Gitartha-Engine/internal/data"
)

func NewRouter(store *data.Store) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	h := &handler{store: store}

	r.GET("/healthz", h.health)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/chapters", h.listChapters)
		v1.GET("/chapters/:chapter", h.getChapter)
		v1.GET("/chapters/:chapter/verses", h.listChapterVerses)
		v1.GET("/chapters/:chapter/verses/:verse", h.getVerse)
		v1.GET("/search", h.searchVerses)
		v1.GET("/random", h.randomVerse)
	}

	return r
}

type handler struct {
	store *data.Store
}

func (h *handler) health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.store.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	})
}

func (h *handler) listChapters(c *gin.Context) {
	chapters, err := h.store.ListChapters(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"chapters": chapters})
}

func (h *handler) getChapter(c *gin.Context) {
	chapterNumber, err := parsePositiveInt(c.Param("chapter"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	chapter, verses, err := h.store.GetChapterWithVerses(c.Request.Context(), chapterNumber)
	if err != nil {
		h.handleDataError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"chapter": chapter,
		"verses":  verses,
	})
}

func (h *handler) listChapterVerses(c *gin.Context) {
	chapterNumber, err := parsePositiveInt(c.Param("chapter"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	_, verses, err := h.store.GetChapterWithVerses(c.Request.Context(), chapterNumber)
	if err != nil {
		h.handleDataError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"chapter": chapterNumber, "verses": verses})
}

func (h *handler) getVerse(c *gin.Context) {
	chapterNumber, err := parsePositiveInt(c.Param("chapter"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chapter number"})
		return
	}

	verseNumber, err := parsePositiveInt(c.Param("verse"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid verse number"})
		return
	}

	verse, err := h.store.GetVerse(c.Request.Context(), chapterNumber, verseNumber)
	if err != nil {
		h.handleDataError(c, err)
		return
	}

	c.JSON(http.StatusOK, verse)
}

func (h *handler) searchVerses(c *gin.Context) {
	query := c.Query("query")
	if len(query) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query must be at least 2 characters"})
		return
	}

	lang := c.DefaultQuery("lang", "en")
	if lang != "en" && lang != "hi" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lang must be 'en' or 'hi'"})
		return
	}

	limitParam := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}

	verses, err := h.store.SearchVerses(c.Request.Context(), query, lang, limit)
	if err != nil {
		h.internalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"results": verses})
}

func (h *handler) randomVerse(c *gin.Context) {
	verse, err := h.store.RandomVerse(c.Request.Context())
	if err != nil {
		h.handleDataError(c, err)
		return
	}

	c.JSON(http.StatusOK, verse)
}

func (h *handler) handleDataError(c *gin.Context, err error) {
	if errors.Is(err, data.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	h.internalError(c, err)
}

func (h *handler) internalError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}

func parsePositiveInt(value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, fmt.Errorf("value must be positive")
	}
	return n, nil
}
