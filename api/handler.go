package api

import (
	"net/http"
	"strconv"

	"github.com/HeisenbergV/repoinsight/pkg/models"
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

	var repositories []models.Repository
	var total int64

	h.db.Model(&models.Repository{}).Count(&total)
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

	var repository models.Repository
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
	// 获取查询参数
	keyword := c.Query("keyword")
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")

	// 转换分页参数
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 limit 参数"})
		return
	}
	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 offset 参数"})
		return
	}

	// 构建查询条件
	query := h.db.Model(&models.Repository{})
	if keyword != "" {
		query = query.Where("full_name LIKE ? OR description LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询总数失败"})
		return
	}

	// 查询数据
	var repositories []models.Repository
	if err := query.
		Select("full_name, created_at, updated_at, ai_analysis, url").
		Order("updated_at DESC").
		Limit(limitInt).
		Offset(offsetInt).
		Find(&repositories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询仓库失败"})
		return
	}

	// 构建响应
	response := gin.H{
		"total":        total,
		"repositories": repositories,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary 获取系统状态
// @Description 获取系统运行状态和统计信息
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} SystemStatus
// @Router /api/v1/system/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	var totalRepos int64
	var lastUpdated models.Repository

	h.db.Model(&models.Repository{}).Count(&totalRepos)
	h.db.Order("updated_at desc").First(&lastUpdated)

	c.JSON(http.StatusOK, gin.H{
		"total_repositories": totalRepos,
		"last_updated":       lastUpdated.UpdatedAt,
		"status":             "running",
	})
}
