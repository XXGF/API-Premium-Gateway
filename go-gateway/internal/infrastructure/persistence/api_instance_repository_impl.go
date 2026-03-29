package persistence

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
)

// ApiInstanceRepositoryImpl API实例仓储实现
type ApiInstanceRepositoryImpl struct {
	db *gorm.DB
}

// NewApiInstanceRepositoryImpl 创建API实例仓储实现
func NewApiInstanceRepositoryImpl(db *gorm.DB) *ApiInstanceRepositoryImpl {
	return &ApiInstanceRepositoryImpl{db: db}
}

func (r *ApiInstanceRepositoryImpl) Insert(instance *entity.ApiInstanceEntity) error {
	if instance.ID == "" {
		instance.ID = uuid.New().String()
	}
	return r.db.Create(instance).Error
}

func (r *ApiInstanceRepositoryImpl) SelectByID(id string) (*entity.ApiInstanceEntity, error) {
	var instance entity.ApiInstanceEntity
	result := r.db.Where("id = ?", id).First(&instance)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &instance, nil
}

func (r *ApiInstanceRepositoryImpl) UpdateByID(instance *entity.ApiInstanceEntity) error {
	return r.db.Save(instance).Error
}

func (r *ApiInstanceRepositoryImpl) DeleteByID(id string) (int64, error) {
	result := r.db.Where("id = ?", id).Delete(&entity.ApiInstanceEntity{})
	return result.RowsAffected, result.Error
}

func (r *ApiInstanceRepositoryImpl) SelectByProjectID(projectID string) ([]*entity.ApiInstanceEntity, error) {
	var instances []*entity.ApiInstanceEntity
	result := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&instances)
	return instances, result.Error
}

func (r *ApiInstanceRepositoryImpl) SelectByProjectIDAndBusinessID(projectID, businessID string) (*entity.ApiInstanceEntity, error) {
	var instance entity.ApiInstanceEntity
	result := r.db.Where("project_id = ? AND business_id = ?", projectID, businessID).First(&instance)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &instance, nil
}

func (r *ApiInstanceRepositoryImpl) SelectByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (*entity.ApiInstanceEntity, error) {
	var instance entity.ApiInstanceEntity
	result := r.db.Where("project_id = ? AND api_type = ? AND business_id = ?", projectID, apiType, businessID).First(&instance)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &instance, nil
}

func (r *ApiInstanceRepositoryImpl) CountByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (int64, error) {
	var count int64
	result := r.db.Model(&entity.ApiInstanceEntity{}).
		Where("project_id = ? AND api_type = ? AND business_id = ?", projectID, apiType, businessID).
		Count(&count)
	return count, result.Error
}

func (r *ApiInstanceRepositoryImpl) SelectByProjectIDAndApiTypeAndStatus(projectID string, apiType entity.ApiType, status entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error) {
	var instances []*entity.ApiInstanceEntity
	result := r.db.Where("project_id = ? AND api_type = ? AND status = ?", projectID, apiType, status).
		Order("created_at DESC").Find(&instances)
	return instances, result.Error
}

func (r *ApiInstanceRepositoryImpl) SelectCandidates(projectID string, apiType entity.ApiType, apiIdentifier string, userID string) ([]*entity.ApiInstanceEntity, error) {
	var instances []*entity.ApiInstanceEntity
	query := r.db.Where("project_id = ? AND api_type = ? AND status = ?", projectID, apiType, entity.ApiInstanceStatusActive).
		Where("(api_identifier = ? OR business_id = ?)", apiIdentifier, apiIdentifier)

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	result := query.Find(&instances)
	return instances, result.Error
}

func (r *ApiInstanceRepositoryImpl) SelectByStatus(status entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error) {
	var instances []*entity.ApiInstanceEntity
	result := r.db.Where("status = ?", status).Order("created_at DESC").Find(&instances)
	return instances, result.Error
}

func (r *ApiInstanceRepositoryImpl) SelectAllWithFilter(projectID string, status *entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error) {
	var instances []*entity.ApiInstanceEntity
	query := r.db.Model(&entity.ApiInstanceEntity{})

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	result := query.Order("created_at DESC").Find(&instances)
	return instances, result.Error
}

func (r *ApiInstanceRepositoryImpl) DeleteByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (int64, error) {
	result := r.db.Where("project_id = ? AND api_type = ? AND business_id = ?", projectID, apiType, businessID).
		Delete(&entity.ApiInstanceEntity{})
	return result.RowsAffected, result.Error
}

func (r *ApiInstanceRepositoryImpl) BatchDeleteByBusinessKeys(projectID string, apiType entity.ApiType, businessIDs []string) (int64, error) {
	result := r.db.Where("project_id = ? AND api_type = ? AND business_id IN ?", projectID, apiType, businessIDs).
		Delete(&entity.ApiInstanceEntity{})
	return result.RowsAffected, result.Error
}
