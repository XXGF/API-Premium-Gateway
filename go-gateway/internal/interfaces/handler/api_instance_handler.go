package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	apiInstanceReq "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request/api_instance"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/response"
)

// ApiInstanceHandler API实例管理处理器（管理接口）
type ApiInstanceHandler struct {
	apiInstanceAppService *appService.ApiInstanceAppService
}

// NewApiInstanceHandler 创建API实例管理处理器
func NewApiInstanceHandler(instanceService *appService.ApiInstanceAppService) *ApiInstanceHandler {
	return &ApiInstanceHandler{apiInstanceAppService: instanceService}
}

// CreateApiInstance 创建API实例
// POST /api/external/api-instances/:projectId
func (h *ApiInstanceHandler) CreateApiInstance(c *gin.Context) {
	projectID := c.Param("projectId")
	var req apiInstanceReq.ApiInstanceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail("请求参数校验失败: "+err.Error()))
		return
	}

	result, err := h.apiInstanceAppService.CreateApiInstance(&req, projectID)
	if err != nil {
		log.Error().Err(err).Msg("创建API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// BatchCreateApiInstances 批量创建API实例
// POST /api/external/api-instances/:projectId/batch
func (h *ApiInstanceHandler) BatchCreateApiInstances(c *gin.Context) {
	projectID := c.Param("projectId")
	var req apiInstanceReq.ApiInstanceBatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail("请求参数校验失败: "+err.Error()))
		return
	}

	result, err := h.apiInstanceAppService.BatchCreateApiInstances(req.Instances, projectID)
	if err != nil {
		log.Error().Err(err).Msg("批量创建API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetApiInstanceById 根据ID获取API实例详情
// GET /api/external/api-instances/detail/:id
func (h *ApiInstanceHandler) GetApiInstanceById(c *gin.Context) {
	id := c.Param("id")
	result, err := h.apiInstanceAppService.GetApiInstanceById(id)
	if err != nil {
		log.Error().Err(err).Msg("获取API实例详情失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetApiInstancesByProjectId 根据项目ID获取API实例列表
// GET /api/external/api-instances/:projectId
func (h *ApiInstanceHandler) GetApiInstancesByProjectId(c *gin.Context) {
	projectID := c.Param("projectId")
	result, err := h.apiInstanceAppService.GetApiInstancesByProjectId(projectID)
	if err != nil {
		log.Error().Err(err).Msg("获取API实例列表失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// UpdateApiInstance 更新API实例
// PUT /api/external/api-instances/:projectId/:apiType/:businessId
func (h *ApiInstanceHandler) UpdateApiInstance(c *gin.Context) {
	projectID := c.Param("projectId")
	apiType := c.Param("apiType")
	businessID := c.Param("businessId")

	var req apiInstanceReq.ApiInstanceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail("请求参数校验失败: "+err.Error()))
		return
	}

	result, err := h.apiInstanceAppService.UpdateApiInstance(projectID, apiType, businessID, &req)
	if err != nil {
		log.Error().Err(err).Msg("更新API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// DeleteApiInstance 删除API实例
// DELETE /api/external/api-instances/:projectId/:apiType/:businessId
func (h *ApiInstanceHandler) DeleteApiInstance(c *gin.Context) {
	projectID := c.Param("projectId")
	apiType := c.Param("apiType")
	businessID := c.Param("businessId")

	at, err := entity.ApiTypeFromCode(apiType)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Fail("无效的 API 类型: "+apiType))
		return
	}

	if err := h.apiInstanceAppService.DeleteApiInstance(projectID, businessID, at); err != nil {
		log.Error().Err(err).Msg("删除API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

// BatchDeleteApiInstances 批量删除API实例
// DELETE /api/external/api-instances/:projectId/batch
func (h *ApiInstanceHandler) BatchDeleteApiInstances(c *gin.Context) {
	projectID := c.Param("projectId")
	var req apiInstanceReq.ApiInstanceBatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Fail("请求参数校验失败: "+err.Error()))
		return
	}

	deleted, err := h.apiInstanceAppService.BatchDeleteApiInstances(projectID, req.Instances)
	if err != nil {
		log.Error().Err(err).Msg("批量删除API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"deleted": deleted}))
}

// ActivateApiInstance 激活API实例
// PUT /api/external/api-instances/:projectId/:apiType/:businessId/activate
func (h *ApiInstanceHandler) ActivateApiInstance(c *gin.Context) {
	projectID := c.Param("projectId")
	apiType := c.Param("apiType")
	businessID := c.Param("businessId")

	result, err := h.apiInstanceAppService.ActivateApiInstance(projectID, apiType, businessID)
	if err != nil {
		log.Error().Err(err).Msg("激活API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// DeactivateApiInstance 停用API实例
// PUT /api/external/api-instances/:projectId/:apiType/:businessId/deactivate
func (h *ApiInstanceHandler) DeactivateApiInstance(c *gin.Context) {
	projectID := c.Param("projectId")
	apiType := c.Param("apiType")
	businessID := c.Param("businessId")

	result, err := h.apiInstanceAppService.DeactivateApiInstance(projectID, apiType, businessID)
	if err != nil {
		log.Error().Err(err).Msg("停用API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// GetAllInstances 获取所有API实例（管理接口）
// GET /api/admin/api-instances
func (h *ApiInstanceHandler) GetAllInstances(c *gin.Context) {
	projectID := c.Query("projectId")
	statusStr := c.Query("status")

	var status *entity.ApiInstanceStatus
	if statusStr != "" {
		s := entity.ApiInstanceStatus(statusStr)
		status = &s
	}

	result, err := h.apiInstanceAppService.GetAllInstancesWithProjects(projectID, status)
	if err != nil {
		log.Error().Err(err).Msg("获取所有API实例失败")
		c.JSON(http.StatusInternalServerError, response.FailWithError(err))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}
