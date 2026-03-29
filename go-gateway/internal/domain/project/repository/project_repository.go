package repository

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/entity"
)

// ProjectRepository 项目仓储接口
type ProjectRepository interface {
	// Insert 插入项目
	Insert(project *entity.ProjectEntity) error
	// SelectByID 根据ID查询项目
	SelectByID(id string) (*entity.ProjectEntity, error)
	// SelectByName 根据名称查询项目
	SelectByName(name string) (*entity.ProjectEntity, error)
	// SelectByApiKey 根据API Key查询项目
	SelectByApiKey(apiKey string) (*entity.ProjectEntity, error)
	// SelectAll 查询所有项目（按创建时间降序）
	SelectAll() ([]*entity.ProjectEntity, error)
	// SelectByStatus 根据状态查询项目
	SelectByStatus(status string) ([]*entity.ProjectEntity, error)
	// SearchByName 按名称模糊搜索项目
	SearchByName(name string) ([]*entity.ProjectEntity, error)
	// CountByName 根据名称统计项目数量
	CountByName(name string) (int64, error)
	// DeleteByID 根据ID删除项目
	DeleteByID(id string) (int64, error)
}
