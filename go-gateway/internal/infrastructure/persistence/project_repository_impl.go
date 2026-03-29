package persistence

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/entity"
)

// ProjectRepositoryImpl 项目仓储实现
type ProjectRepositoryImpl struct {
	db *gorm.DB
}

// NewProjectRepositoryImpl 创建项目仓储实现
func NewProjectRepositoryImpl(db *gorm.DB) *ProjectRepositoryImpl {
	return &ProjectRepositoryImpl{db: db}
}

func (r *ProjectRepositoryImpl) Insert(project *entity.ProjectEntity) error {
	if project.ID == "" {
		project.ID = uuid.New().String()
	}
	return r.db.Create(project).Error
}

func (r *ProjectRepositoryImpl) SelectByID(id string) (*entity.ProjectEntity, error) {
	var project entity.ProjectEntity
	result := r.db.Where("id = ?", id).First(&project)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &project, nil
}

func (r *ProjectRepositoryImpl) SelectByName(name string) (*entity.ProjectEntity, error) {
	var project entity.ProjectEntity
	result := r.db.Where("name = ?", name).First(&project)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &project, nil
}

func (r *ProjectRepositoryImpl) SelectByApiKey(apiKey string) (*entity.ProjectEntity, error) {
	var project entity.ProjectEntity
	result := r.db.Where("api_key = ?", apiKey).First(&project)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &project, nil
}

func (r *ProjectRepositoryImpl) SelectAll() ([]*entity.ProjectEntity, error) {
	var projects []*entity.ProjectEntity
	result := r.db.Order("created_at DESC").Find(&projects)
	return projects, result.Error
}

func (r *ProjectRepositoryImpl) SelectByStatus(status string) ([]*entity.ProjectEntity, error) {
	var projects []*entity.ProjectEntity
	result := r.db.Where("status = ?", status).Order("created_at DESC").Find(&projects)
	return projects, result.Error
}

func (r *ProjectRepositoryImpl) SearchByName(name string) ([]*entity.ProjectEntity, error) {
	var projects []*entity.ProjectEntity
	result := r.db.Where("name LIKE ?", "%"+name+"%").Order("created_at DESC").Find(&projects)
	return projects, result.Error
}

func (r *ProjectRepositoryImpl) CountByName(name string) (int64, error) {
	var count int64
	result := r.db.Model(&entity.ProjectEntity{}).Where("name = ?", name).Count(&count)
	return count, result.Error
}

func (r *ProjectRepositoryImpl) DeleteByID(id string) (int64, error) {
	result := r.db.Where("id = ?", id).Delete(&entity.ProjectEntity{})
	return result.RowsAffected, result.Error
}
