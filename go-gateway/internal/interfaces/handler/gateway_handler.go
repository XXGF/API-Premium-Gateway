// Package handler 实现了 HTTP 请求处理器。
//
// 该包属于 DDD 架构的接口层，负责接收 HTTP 请求、参数校验、
// 调用应用层服务并返回统一格式的响应。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/response"
)

// GatewayHandler 网关核心接口处理器。
//
// 提供 API 实例选择和调用结果上报两个核心接口，
// 是上游服务与 Gateway 交互的主要入口。
type GatewayHandler struct {
	selectionAppService *appService.SelectionAppService
}

// NewGatewayHandler 创建网关处理器
func NewGatewayHandler(selectionService *appService.SelectionAppService) *GatewayHandler {
	return &GatewayHandler{selectionAppService: selectionService}
}

// SelectInstance 选择最佳 API 实例。
//
// POST /api/external/gateway/select?projectId={projectId}
//
// 根据负载均衡策略从候选实例中选择当前最佳的 API 实例。
// 支持降级链（Fallback Chain），当主要实例不可用时自动尝试备选。
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

// ReportResult 上报调用结果。
//
// POST /api/external/gateway/report?projectId={projectId}
//
// 上游服务调用后端 API 后，将调用结果（成功/失败、延迟、错误信息等）
// 上报给 Gateway，用于更新指标和健康状态。
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

// HealthCheck 健康检查接口。
//
// GET /api/health
//
// 无需认证，用于 Docker HEALTHCHECK 和负载均衡器探测。
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, response.Success(gin.H{
		"status":  "UP",
		"service": "API-Premium-Gateway (Go)",
	}))
}
