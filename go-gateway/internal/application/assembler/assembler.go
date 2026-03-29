package assembler

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/dto"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/command"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	metricsCommand "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/command"
	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request"
	apiInstanceReq "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request/api_instance"
)

// ApiInstanceToDTO 将实体转换为DTO
func ApiInstanceToDTO(e *entity.ApiInstanceEntity) *dto.ApiInstanceDTO {
	if e == nil {
		return nil
	}
	return &dto.ApiInstanceDTO{
		ID:            e.ID,
		ProjectID:     e.ProjectID,
		UserID:        e.UserID,
		ApiIdentifier: e.ApiIdentifier,
		ApiType:       string(e.ApiType),
		BusinessID:    e.BusinessID,
		RoutingParams: metricsEntity.JSONBMap(e.RoutingParams),
		Status:        string(e.Status),
		Metadata:      metricsEntity.JSONBMap(e.Metadata),
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

// ApiInstanceToDTOList 将实体列表转换为DTO列表
func ApiInstanceToDTOList(entities []*entity.ApiInstanceEntity) []*dto.ApiInstanceDTO {
	var dtos []*dto.ApiInstanceDTO
	for _, e := range entities {
		dtos = append(dtos, ApiInstanceToDTO(e))
	}
	return dtos
}

// CreateRequestToEntity 将创建请求转换为实体
func CreateRequestToEntity(req *apiInstanceReq.ApiInstanceCreateRequest, projectID string) *entity.ApiInstanceEntity {
	return &entity.ApiInstanceEntity{
		ProjectID:     projectID,
		UserID:        req.UserID,
		ApiIdentifier: req.ApiIdentifier,
		ApiType:       entity.ApiType(req.ApiType),
		BusinessID:    req.BusinessID,
		RoutingParams: metricsEntity.JSONBMap(req.RoutingParams),
		Status:        entity.ApiInstanceStatusActive,
		Metadata:      metricsEntity.JSONBMap(req.Metadata),
	}
}

// CreateRequestToEntityList 将创建请求列表转换为实体列表
func CreateRequestToEntityList(reqs []*apiInstanceReq.ApiInstanceCreateRequest, projectID string) []*entity.ApiInstanceEntity {
	var entities []*entity.ApiInstanceEntity
	for _, req := range reqs {
		entities = append(entities, CreateRequestToEntity(req, projectID))
	}
	return entities
}

// UpdateRequestToEntity 将更新请求转换为实体
func UpdateRequestToEntity(req *apiInstanceReq.ApiInstanceUpdateRequest, projectID string) *entity.ApiInstanceEntity {
	return &entity.ApiInstanceEntity{
		ProjectID:     projectID,
		UserID:        req.UserID,
		ApiIdentifier: req.ApiIdentifier,
		RoutingParams: metricsEntity.JSONBMap(req.RoutingParams),
		Metadata:      metricsEntity.JSONBMap(req.Metadata),
	}
}

// SelectRequestToCommand 将选择请求转换为命令
func SelectRequestToCommand(req *request.SelectInstanceRequest, projectID string) *command.InstanceSelectionCommand {
	cmd := command.NewInstanceSelectionCommand(projectID, req.UserID, req.ApiIdentifier, req.ApiType)

	// 设置亲和性上下文
	if req.HasAffinityRequirement() {
		cmd.AffinityContext = &entity.AffinityContext{
			AffinityType: req.AffinityType,
			AffinityKey:  req.AffinityKey,
			Strength:     entity.AffinityStrengthPreferred,
		}
	}

	return cmd
}

// ReportRequestToCommand 将上报请求转换为命令
func ReportRequestToCommand(req *request.ReportResultRequest, projectID string) *metricsCommand.CallResultCommand {
	return &metricsCommand.CallResultCommand{
		ProjectID:     projectID,
		InstanceID:    req.InstanceID,
		Success:       req.Success,
		LatencyMs:     req.LatencyMs,
		ErrorMessage:  req.ErrorMessage,
		ErrorType:     req.ErrorType,
		UsageMetrics:  req.UsageMetrics,
		CallTimestamp: req.CallTimestamp,
	}
}
