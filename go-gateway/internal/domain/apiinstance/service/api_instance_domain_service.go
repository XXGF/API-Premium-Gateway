package service

import (
	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/repository"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/exception"
)

// ApiInstanceDomainService API实例领域服务
type ApiInstanceDomainService struct {
	apiInstanceRepository repository.ApiInstanceRepository
}

// NewApiInstanceDomainService 创建API实例领域服务
func NewApiInstanceDomainService(repo repository.ApiInstanceRepository) *ApiInstanceDomainService {
	return &ApiInstanceDomainService{apiInstanceRepository: repo}
}

// CreateApiInstance 创建API实例
func (s *ApiInstanceDomainService) CreateApiInstance(instance *entity.ApiInstanceEntity) (*entity.ApiInstanceEntity, error) {
	// 检查是否已存在
	exists, err := s.IsApiInstanceExists(instance.ProjectID, instance.ApiType, instance.BusinessID)
	if err != nil {
		return nil, err
	}
	if exists {
		log.Info().Str("projectId", instance.ProjectID).Str("businessId", instance.BusinessID).Msg("API实例已存在，返回已存在的实例")
		return s.GetApiInstanceByBusinessKey(instance.ProjectID, instance.ApiType, instance.BusinessID)
	}

	if err := s.apiInstanceRepository.Insert(instance); err != nil {
		return nil, err
	}
	log.Info().Str("id", instance.ID).Str("businessId", instance.BusinessID).Msg("API实例创建成功")
	return instance, nil
}

// BatchCreateApiInstances 批量创建API实例
func (s *ApiInstanceDomainService) BatchCreateApiInstances(instances []*entity.ApiInstanceEntity) ([]*entity.ApiInstanceEntity, error) {
	if len(instances) == 0 {
		return nil, nil
	}

	log.Info().Int("count", len(instances)).Msg("开始批量创建API实例")

	var created []*entity.ApiInstanceEntity
	for _, instance := range instances {
		exists, err := s.IsApiInstanceExists(instance.ProjectID, instance.ApiType, instance.BusinessID)
		if err != nil {
			return nil, err
		}
		if exists {
			log.Info().Str("businessId", instance.BusinessID).Msg("API实例已存在，跳过创建")
			continue
		}
		if err := s.apiInstanceRepository.Insert(instance); err != nil {
			return nil, err
		}
		created = append(created, instance)
	}

	log.Info().Int("created", len(created)).Int("skipped", len(instances)-len(created)).Msg("批量创建API实例完成")
	return created, nil
}

// IsApiInstanceExists 检查API实例是否已存在
func (s *ApiInstanceDomainService) IsApiInstanceExists(projectID string, apiType entity.ApiType, businessID string) (bool, error) {
	count, err := s.apiInstanceRepository.CountByBusinessKey(projectID, apiType, businessID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetApiInstanceById 根据ID获取API实例详情
func (s *ApiInstanceDomainService) GetApiInstanceById(id string) (*entity.ApiInstanceEntity, error) {
	instance, err := s.apiInstanceRepository.SelectByID(id)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, exception.NewEntityNotFoundError("API实例不存在，ID: " + id)
	}
	return instance, nil
}

// GetApiInstancesByProjectId 根据项目ID获取API实例列表
func (s *ApiInstanceDomainService) GetApiInstancesByProjectId(projectID string) ([]*entity.ApiInstanceEntity, error) {
	return s.apiInstanceRepository.SelectByProjectID(projectID)
}

// GetApiInstanceByBusinessId 根据业务ID和项目ID获取API实例
func (s *ApiInstanceDomainService) GetApiInstanceByBusinessId(projectID, businessID string) (*entity.ApiInstanceEntity, error) {
	instance, err := s.apiInstanceRepository.SelectByProjectIDAndBusinessID(projectID, businessID)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, exception.NewEntityNotFoundError("API实例不存在，项目ID: " + projectID + ", 业务ID: " + businessID)
	}
	return instance, nil
}

// GetApiInstanceByBusinessKey 根据业务键获取API实例
func (s *ApiInstanceDomainService) GetApiInstanceByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (*entity.ApiInstanceEntity, error) {
	instance, err := s.apiInstanceRepository.SelectByBusinessKey(projectID, apiType, businessID)
	if err != nil {
		return nil, err
	}
	if instance == nil {
		return nil, exception.NewEntityNotFoundError("API实例不存在，projectId=" + projectID + ", apiType=" + string(apiType) + ", businessId=" + businessID)
	}
	return instance, nil
}

// GetApiInstancesByStatus 获取指定状态的API实例列表
func (s *ApiInstanceDomainService) GetApiInstancesByStatus(status entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error) {
	return s.apiInstanceRepository.SelectByStatus(status)
}

// GetApiInstancesByApiType 根据API类型获取API实例列表
func (s *ApiInstanceDomainService) GetApiInstancesByApiType(projectID string, apiType entity.ApiType) ([]*entity.ApiInstanceEntity, error) {
	return s.apiInstanceRepository.SelectByProjectIDAndApiTypeAndStatus(projectID, apiType, entity.ApiInstanceStatusActive)
}

// UpdateApiInstance 更新API实例
func (s *ApiInstanceDomainService) UpdateApiInstance(instance *entity.ApiInstanceEntity) (*entity.ApiInstanceEntity, error) {
	if err := s.apiInstanceRepository.UpdateByID(instance); err != nil {
		return nil, err
	}
	log.Info().Str("id", instance.ID).Msg("API实例更新成功")
	return instance, nil
}

// DeleteApiInstance 删除API实例
func (s *ApiInstanceDomainService) DeleteApiInstance(id string) (bool, error) {
	deleted, err := s.apiInstanceRepository.DeleteByID(id)
	if err != nil {
		return false, err
	}
	if deleted > 0 {
		log.Info().Str("id", id).Msg("API实例删除成功")
	}
	return deleted > 0, nil
}

// DeleteApiInstanceByBusinessKey 根据业务键删除API实例
func (s *ApiInstanceDomainService) DeleteApiInstanceByBusinessKey(projectID, businessID string, apiType entity.ApiType) error {
	_, err := s.apiInstanceRepository.DeleteByBusinessKey(projectID, apiType, businessID)
	return err
}

// GetAllInstancesWithProjects 获取所有API实例
func (s *ApiInstanceDomainService) GetAllInstancesWithProjects(projectID string, status *entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error) {
	return s.apiInstanceRepository.SelectAllWithFilter(projectID, status)
}

// BatchDeleteApiInstances 批量删除API实例
func (s *ApiInstanceDomainService) BatchDeleteApiInstances(projectID string, deleteKeys []ApiInstanceDeleteKey) (int, error) {
	if len(deleteKeys) == 0 {
		return 0, nil
	}

	log.Info().Str("projectId", projectID).Int("count", len(deleteKeys)).Msg("开始批量删除API实例")

	// 按 apiType 分组
	grouped := make(map[entity.ApiType][]string)
	for _, key := range deleteKeys {
		grouped[key.ApiType] = append(grouped[key.ApiType], key.BusinessID)
	}

	totalDeleted := 0
	for apiType, businessIDs := range grouped {
		deleted, err := s.apiInstanceRepository.BatchDeleteByBusinessKeys(projectID, apiType, businessIDs)
		if err != nil {
			log.Error().Str("apiType", string(apiType)).Err(err).Msg("批量删除API实例失败")
			continue
		}
		totalDeleted += int(deleted)
	}

	log.Info().Int("deleted", totalDeleted).Msg("批量删除API实例完成")
	return totalDeleted, nil
}

// ApiInstanceDeleteKey API实例删除键
type ApiInstanceDeleteKey struct {
	ApiType    entity.ApiType
	BusinessID string
}
