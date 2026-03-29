package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/response"
)

// GatewayHandler 网关处理器（对外接口）
type GatewayHandler struct {
	selectionAppService *appService.SelectionAppService
}

// NewGatewayHandler 创建网关处理器
func NewGatewayHandler(selectionService *appService.SelectionAppService) *GatewayHandler {
	return &GatewayHandler{selectionAppService: selectionService}
}

// SelectInstance 选择最佳API实例
// POST /api/external/gateway/select
func (h *GatewayHandler) SelectInstance(c *gin.Context) {
	var req request.SelectInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("请求参数校验失败")
		c.JSON(http.StatusBadRequest, response.Fail("请求参数校验失败: "+err.Error()))
		return
	}

	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, response.Fail("缺少 projectId 参数"))
		return
	}

	result, err := h.selectionAppService.SelectBestInstance(&req, projectID)
	if err != nil {
		log.Error().Err(err).Msg("选择API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// ReportResult 上报调用结果
// POST /api/external/gateway/report
func (h *GatewayHandler) ReportResult(c *gin.Context) {
	var req request.ReportResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("请求参数校验失败")
		c.JSON(http.StatusBadRequest, response.Fail("请求参数校验失败: "+err.Error()))
		return
	}

	projectID := c.Query("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, response.Fail("缺少 projectId 参数"))
		return
	}

	if err := h.selectionAppService.ReportCallResult(&req, projectID); err != nil {
		log.Error().Err(err).Msg("上报调用结果失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// HealthCheck 健康检查
// GET /api/health
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(gin.H{
		"status":  "UP",
		"service": "API-Premium-Gateway (Go)",
	}))
}
