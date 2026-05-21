package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/manuelRestrep0/watchdog/model"
	"github.com/manuelRestrep0/watchdog/monitor"
	"github.com/manuelRestrep0/watchdog/store"
)

type TargetHandler struct {
	store   *store.SQLiteStore
	monitor *monitor.Monitor
	redis   *store.RedisStore
}

func NewTargetHandler(s *store.SQLiteStore, m *monitor.Monitor, r *store.RedisStore) *TargetHandler {
	return &TargetHandler{
		store:   s,
		monitor: m,
		redis:   r,
	}
}

func (h *TargetHandler) Create(c *gin.Context) {
	var body struct {
		URL      string `json:"url"      binding:"required"`
		Interval int    `json:"interval" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target, err := h.store.CreateTarget(body.URL, body.Interval)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.monitor.Start(*target)
	c.JSON(http.StatusCreated, target)
}

func (h *TargetHandler) List(c *gin.Context) {
	targets, err := h.store.ListTargets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, targets)
}

func (h *TargetHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	h.monitor.Stop(id)

	if err := h.store.DeleteTarget(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func (h *TargetHandler) History(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	checks, err := h.store.GetHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, checks)
}

func (h *TargetHandler) StartExisting(targets []model.Target) {
	for _, t := range targets {
		h.monitor.Start(t)
	}
}

func (h *TargetHandler) LastCheck(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	check, err := h.redis.GetLastCheck(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if check == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no checks yet"})
		return
	}

	c.JSON(http.StatusOK, check)
}

func (h *TargetHandler) RegisterRoutes(r *gin.Engine) {
	targets := r.Group("/targets")
	targets.POST("", h.Create)
	targets.GET("", h.List)
	targets.DELETE("/:id", h.Delete)
	targets.GET("/:id/history", h.History)
	targets.GET("/:id/last", h.LastCheck)
}
