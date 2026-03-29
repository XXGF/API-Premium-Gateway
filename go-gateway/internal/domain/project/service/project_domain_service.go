package service

import (
	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/repository"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/entity"
	projectRepo "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/repository"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/exception"
)

// ProjectDomainService 项目领域服务
type ProjectDomainService struct {
	projectRepository projectRepo.ProjectRepository
	apiKeyRepository  repository.ApiKeyRepository
}

// NewProjectDomainService 创建项目领域服务
func NewProjectDomainService(projectRepo projectRepo.ProjectRepository, apiKeyRepo repository.ApiKeyRepository) *ProjectDomainService {
	return &ProjectDomainService{
		projectRepository: projectRepo,
		apiKeyRepository:  apiKeyRepo,
	}
}

// CreateProject 创建项目
func (s *ProjectDomainService) CreateProject(name, description, apiKey string) (*entity.ProjectEntity, error) {
	log.Info().Str("name", name).Str("apiKey", apiKey).Msg("开始创建项目")

	// 验证apiKey是否存在并且激活
	if err := s.validateApiKey(apiKey); err != nil {
		return nil, err
	}

	// 检查是否已存在使用该apiKey的项目
	existing, err := s.GetProjectByApiKey(apiKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	// 创建项目实体
	project := entity.NewProjectEntity(name, description, apiKey)

	// 保存项目
	if err := s.projectRepository.Insert(project); err != nil {
		return nil, err
	}

	log.Info().Str("id", project.ID).Str("name", project.Name).Msg("项目创建成功")
	return project, nil
}

// GetProjectByApiKey 根据API Key获取项目
func (s *ProjectDomainService) GetProjectByApiKey(apiKey string) (*entity.ProjectEntity, error) {
	return s.projectRepository.SelectByApiKey(apiKey)
}

// GetProjectById 根据ID获取项目详情
func (s *ProjectDomainService) GetProjectById(id string) (*entity.ProjectEntity, error) {
	log.Debug().Str("id", id).Msg("获取项目详情")

	project, err := s.projectRepository.SelectByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, exception.NewEntityNotFoundError("项目不存在，ID: " + id)
	}
	return project, nil
}

// GetProjectList 获取项目列表
func (s *ProjectDomainService) GetProjectList() ([]*entity.ProjectEntity, error) {
	log.Debug().Msg("获取项目列表")
	return s.projectRepository.SelectAll()
}

// ValidateProjectExists 验证项目是否存在
func (s *ProjectDomainService) ValidateProjectExists(projectId string) error {
	project, err := s.projectRepository.SelectByID(projectId)
	if err != nil {
		return err
	}
	if project == nil {
		return exception.NewEntityNotFoundError("项目不存在，ID: " + projectId)
	}
	return nil
}

// IsProjectActive 检查项目是否存在且处于活跃状态
func (s *ProjectDomainService) IsProjectActive(projectId string) bool {
	if projectId == "" {
		return false
	}

	project, err := s.projectRepository.SelectByID(projectId)
	if err != nil || project == nil {
		log.Debug().Str("projectId", projectId).Msg("项目不存在")
		return false
	}

	isActive := project.IsActive()
	if !isActive {
		log.Debug().Str("projectId", projectId).Str("status", string(project.Status)).Msg("项目不是活跃状态")
	}
	return isActive
}

// GetAllProjects 获取所有项目
func (s *ProjectDomainService) GetAllProjects() ([]*entity.ProjectEntity, error) {
	log.Debug().Msg("获取所有项目列表")
	return s.projectRepository.SelectAll()
}

// SearchProjectsByName 根据项目名称搜索项目
func (s *ProjectDomainService) SearchProjectsByName(projectName string) ([]*entity.ProjectEntity, error) {
	log.Debug().Str("name", projectName).Msg("按名称搜索项目")
	return s.projectRepository.SearchByName(projectName)
}

// GetProjectsByStatus 根据状态获取项目列表
func (s *ProjectDomainService) GetProjectsByStatus(status string) ([]*entity.ProjectEntity, error) {
	log.Debug().Str("status", status).Msg("按状态查询项目")
	return s.projectRepository.SelectByStatus(status)
}

// DeleteProject 删除项目
func (s *ProjectDomainService) DeleteProject(projectId string) error {
	log.Warn().Str("projectId", projectId).Msg("删除项目")

	if err := s.ValidateProjectExists(projectId); err != nil {
		return err
	}

	deleted, err := s.projectRepository.DeleteByID(projectId)
	if err != nil {
		return err
	}
	if deleted > 0 {
		log.Warn().Str("projectId", projectId).Msg("项目删除成功")
	} else {
		return exception.NewBusinessErrorWithCode("PROJECT_DELETE_FAILED", "项目删除失败，项目ID: "+projectId)
	}
	return nil
}

// GetProjectIdByApiKey 根据API Key获取项目ID
func (s *ProjectDomainService) GetProjectIdByApiKey(apiKeyValue string) (string, error) {
	log.Debug().Str("apiKey", apiKeyValue).Msg("根据API Key查找项目ID")

	project, err := s.projectRepository.SelectByApiKey(apiKeyValue)
	if err != nil {
		return "", err
	}
	if project == nil {
		log.Debug().Str("apiKey", apiKeyValue).Msg("未找到使用此API Key的项目")
		return "", nil
	}

	log.Debug().Str("id", project.ID).Str("name", project.Name).Msg("找到项目")
	return project.ID, nil
}

// GetProjectNameById 根据项目ID获取项目名称
func (s *ProjectDomainService) GetProjectNameById(projectId string) string {
	log.Debug().Str("projectId", projectId).Msg("根据项目ID获取项目名称")

	project, err := s.projectRepository.SelectByID(projectId)
	if err != nil || project == nil {
		log.Debug().Str("projectId", projectId).Msg("项目不存在")
		return ""
	}
	return project.Name
}

// validateApiKey 验证API Key是否存在并且可用
func (s *ProjectDomainService) validateApiKey(apiKey string) error {
	apiKeyEntity, err := s.apiKeyRepository.SelectByApiKeyValue(apiKey)
	if err != nil {
		return err
	}
	if apiKeyEntity == nil {
		return exception.NewBusinessErrorWithCode("INVALID_API_KEY", "API Key 不存在: "+apiKey)
	}
	if !apiKeyEntity.IsUsable() {
		return exception.NewBusinessErrorWithCode("API_KEY_UNUSABLE", "API Key 不可用，状态: "+string(apiKeyEntity.Status))
	}

	log.Debug().Str("apiKey", apiKey).Msg("API Key 验证通过")
	return nil
}
