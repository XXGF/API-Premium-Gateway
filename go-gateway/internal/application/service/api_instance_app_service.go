package service

import (
	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/assembler"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/dto"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	apiInstanceService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/service"
	projectService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/service"
	apiInstanceReq "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request/api_instance"
)

// ApiInstanceAppService API实例应用服务
type ApiInstanceAppService struct {
	apiInstanceDomainService *apiInstanceService.ApiInstanceDomainService
	projectDomainService     *projectService.ProjectDomainService
}

// NewApiInstanceAppService 创建API实例应用服务
func NewApiInstanceAppService(
	instanceService *apiInstanceService.ApiInstanceDomainService,
	projectService *projectService.ProjectDomainService,
) *ApiInstanceAppService {
	return &ApiInstanceAppService{
		apiInstanceDomainService: instanceService,
		projectDomainService:     projectService,
	}
}

// IsApiInstanceExists 检查API实例是否已存在
func (s *ApiInstanceAppService) IsApiInstanceExists(projectID string, apiType entity.ApiType, businessID string) bool {
	exists, _ := s.apiInstanceDomainService.IsApiInstanceExists(projectID, apiType, businessID)
	return exists
}

// CreateApiInstance 创建API实例
func (s *ApiInstanceAppService) CreateApiInstance(req *apiInstanceReq.ApiInstanceCreateRequest, projectID string) (*dto.ApiInstanceDTO, error) {
	if err := s.projectDomainService.ValidateProjectExists(projectID); err != nil {
		return nil, err
	}

	e := assembler.CreateRequestToEntity(req, projectID)
	alreadyExists, _ := s.apiInstanceDomainService.IsApiInstanceExists(projectID, e.ApiType, e.BusinessID)

	created, err := s.apiInstanceDomainService.CreateApiInstance(e)
	if err != nil {
		return nil, err
	}

	if alreadyExists {
		log.Info().Str("id", created.ID).Msg("API实例已存在，返回已存在实例")
	} else {
		log.Info().Str("id", created.ID).Msg("API实例创建成功")
	}

	return assembler.ApiInstanceToDTO(created), nil
}

// BatchCreateApiInstances 批量创建API实例
func (s *ApiInstanceAppService) BatchCreateApiInstances(reqs []*apiInstanceReq.ApiInstanceCreateRequest, projectID string) ([]*dto.ApiInstanceDTO, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	if err := s.projectDomainService.ValidateProjectExists(projectID); err != nil {
		return nil, err
	}

	entities := assembler.CreateRequestToEntityList(reqs, projectID)
	created, err := s.apiInstanceDomainService.BatchCreateApiInstances(entities)
	if err != nil {
		return nil, err
	}

	return assembler.ApiInstanceToDTOList(created), nil
}

// GetApiInstanceById 根据ID获取API实例详情
func (s *ApiInstanceAppService) GetApiInstanceById(id string) (*dto.ApiInstanceDTO, error) {
	e, err := s.apiInstanceDomainService.GetApiInstanceById(id)
	if err != nil {
		return nil, err
	}
	return assembler.ApiInstanceToDTO(e), nil
}

// GetApiInstancesByProjectId 根据项目ID获取API实例列表
func (s *ApiInstanceAppService) GetApiInstancesByProjectId(projectID string) ([]*dto.ApiInstanceDTO, error) {
	entities, err := s.apiInstanceDomainService.GetApiInstancesByProjectId(projectID)
	if err != nil {
		return nil, err
	}
	return assembler.ApiInstanceToDTOList(entities), nil
}

// UpdateApiInstance 更新API实例
func (s *ApiInstanceAppService) UpdateApiInstance(projectID, apiType, businessID string, req *apiInstanceReq.ApiInstanceUpdateRequest) (*dto.ApiInstanceDTO, error) {
	log.Info().Str("projectId", projectID).Str("apiType", apiType).Str("businessId", businessID).Msg("开始更新API实例")

	if err := s.projectDomainService.ValidateProjectExists(projectID); err != nil {
		return nil, err
	}

	at, err := entity.ApiTypeFromCode(apiType)
	if err != nil {
		return nil, err
	}

	existing, err := s.apiInstanceDomainService.GetApiInstanceByBusinessKey(projectID, at, businessID)
	if err != nil {
		return nil, err
	}

	updateEntity := assembler.UpdateRequestToEntity(req, projectID)
	updateEntity.ID = existing.ID
	updateEntity.ApiType = at
	updateEntity.BusinessID = businessID

	updated, err := s.apiInstanceDomainService.UpdateApiInstance(updateEntity)
	if err != nil {
		return nil, err
	}

	return assembler.ApiInstanceToDTO(updated), nil
}

// DeleteApiInstance 删除API实例
func (s *ApiInstanceAppService) DeleteApiInstance(projectID, businessID string, apiType entity.ApiType) error {
	return s.apiInstanceDomainService.DeleteApiInstanceByBusinessKey(projectID, businessID, apiType)
}

// BatchDeleteApiInstances 批量删除API实例
func (s *ApiInstanceAppService) BatchDeleteApiInstances(projectID string, deleteItems []*apiInstanceReq.ApiInstanceDeleteItem) (int, error) {
	if len(deleteItems) == 0 {
		return 0, nil
	}

	if err := s.projectDomainService.ValidateProjectExists(projectID); err != nil {
		return 0, err
	}

	var deleteKeys []apiInstanceService.ApiInstanceDeleteKey
	for _, item := range deleteItems {
		deleteKeys = append(deleteKeys, apiInstanceService.ApiInstanceDeleteKey{
			ApiType:    entity.ApiType(item.ApiType),
			BusinessID: item.BusinessID,
		})
	}

	return s.apiInstanceDomainService.BatchDeleteApiInstances(projectID, deleteKeys)
}

// ActivateApiInstance 激活API实例
func (s *ApiInstanceAppService) ActivateApiInstance(projectID, apiType, businessID string) (*dto.ApiInstanceDTO, error) {
	at, err := entity.ApiTypeFromCode(apiType)
	if err != nil {
		return nil, err
	}
	e, err := s.apiInstanceDomainService.GetApiInstanceByBusinessKey(projectID, at, businessID)
	if err != nil {
		return nil, err
	}
	e.Activate()
	updated, err := s.apiInstanceDomainService.UpdateApiInstance(e)
	if err != nil {
		return nil, err
	}
	return assembler.ApiInstanceToDTO(updated), nil
}

// DeactivateApiInstance 停用API实例
func (s *ApiInstanceAppService) DeactivateApiInstance(projectID, apiType, businessID string) (*dto.ApiInstanceDTO, error) {
	at, err := entity.ApiTypeFromCode(apiType)
	if err != nil {
		return nil, err
	}
	e, err := s.apiInstanceDomainService.GetApiInstanceByBusinessKey(projectID, at, businessID)
	if err != nil {
		return nil, err
	}
	e.Deactivate()
	updated, err := s.apiInstanceDomainService.UpdateApiInstance(e)
	if err != nil {
		return nil, err
	}
	return assembler.ApiInstanceToDTO(updated), nil
}

// DeprecateApiInstance 标记API实例为已弃用
func (s *ApiInstanceAppService) DeprecateApiInstance(projectID, apiType, businessID string) (*dto.ApiInstanceDTO, error) {
	at, err := entity.ApiTypeFromCode(apiType)
	if err != nil {
		return nil, err
	}
	e, err := s.apiInstanceDomainService.GetApiInstanceByBusinessKey(projectID, at, businessID)
	if err != nil {
		return nil, err
	}
	e.Deprecate()
	updated, err := s.apiInstanceDomainService.UpdateApiInstance(e)
	if err != nil {
		return nil, err
	}
	return assembler.ApiInstanceToDTO(updated), nil
}

// GetAllInstancesWithProjects 获取所有API实例（包含项目信息）
func (s *ApiInstanceAppService) GetAllInstancesWithProjects(projectID string, status *entity.ApiInstanceStatus) ([]*dto.ApiInstanceDTO, error) {
	entities, err := s.apiInstanceDomainService.GetAllInstancesWithProjects(projectID, status)
	if err != nil {
		return nil, err
	}
	dtos := assembler.ApiInstanceToDTOList(entities)

	// 填充项目名称
	for _, d := range dtos {
		pName := s.projectDomainService.GetProjectNameById(d.ProjectID)
		if pName == "" {
			d.ProjectName = "未知项目"
		} else {
			d.ProjectName = pName
		}
	}

	return dtos, nil
}
