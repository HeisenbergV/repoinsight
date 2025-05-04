package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// @Summary 获取仓库列表
// @Description 获取所有已分析的仓库列表
// @Tags repositories
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} Response
// @Router /api/v1/repositories [get]
func (h *Handler) GetRepositories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var repositories []Repository
	var total int64

	h.db.Model(&Repository{}).Count(&total)
	h.db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&repositories)

	c.JSON(http.StatusOK, gin.H{
		"data": repositories,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// @Summary 获取仓库详情
// @Description 获取指定仓库的详细信息
// @Tags repositories
// @Accept json
// @Produce json
// @Param id path int true "仓库ID"
// @Success 200 {object} Repository
// @Router /api/v1/repositories/{id} [get]
func (h *Handler) GetRepository(c *gin.Context) {
	id := c.Param("id")

	var repository Repository
	if err := h.db.First(&repository, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	c.JSON(http.StatusOK, repository)
}

// @Summary 搜索仓库
// @Description 根据关键词搜索仓库
// @Tags repositories
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} Response
// @Router /api/v1/repositories/search [get]
func (h *Handler) SearchRepositories(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	var repositories []Repository
	var total int64

	query := h.db.Model(&Repository{}).Where("full_name LIKE ? OR description LIKE ? OR analysis LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	query.Count(&total)
	query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&repositories)

	c.JSON(http.StatusOK, gin.H{
		"data": repositories,
		"meta": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// @Summary 获取系统状态
// @Description 获取系统运行状态和统计信息
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} SystemStatus
// @Router /api/v1/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	var totalRepos int64
	var lastUpdated Repository

	h.db.Model(&Repository{}).Count(&totalRepos)
	h.db.Order("updated_at desc").First(&lastUpdated)

	c.JSON(http.StatusOK, gin.H{
		"total_repositories": totalRepos,
		"last_updated":       lastUpdated.UpdatedAt,
		"status":             "running",
	})
}
