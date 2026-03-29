package repository

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
)

// ApiInstanceRepository API实例仓储接口
type ApiInstanceRepository interface {
	// Insert 插入API实例
	Insert(instance *entity.ApiInstanceEntity) error
	// SelectByID 根据ID查询
	SelectByID(id string) (*entity.ApiInstanceEntity, error)
	// UpdateByID 根据ID更新
	UpdateByID(instance *entity.ApiInstanceEntity) error
	// DeleteByID 根据ID删除
	DeleteByID(id string) (int64, error)
	// SelectByProjectID 根据项目ID查询
	SelectByProjectID(projectID string) ([]*entity.ApiInstanceEntity, error)
	// SelectByProjectIDAndBusinessID 根据项目ID和业务ID查询
	SelectByProjectIDAndBusinessID(projectID, businessID string) (*entity.ApiInstanceEntity, error)
	// SelectByBusinessKey 根据业务键查询（projectId + apiType + businessId）
	SelectByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (*entity.ApiInstanceEntity, error)
	// CountByBusinessKey 根据业务键统计数量
	CountByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (int64, error)
	// SelectByProjectIDAndApiTypeAndStatus 根据项目ID、API类型和状态查询
	SelectByProjectIDAndApiTypeAndStatus(projectID string, apiType entity.ApiType, status entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error)
	// SelectCandidates 查找候选实例
	SelectCandidates(projectID string, apiType entity.ApiType, apiIdentifier string, userID string) ([]*entity.ApiInstanceEntity, error)
	// SelectByStatus 根据状态查询
	SelectByStatus(status entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error)
	// SelectAll 查询所有（支持过滤）
	SelectAllWithFilter(projectID string, status *entity.ApiInstanceStatus) ([]*entity.ApiInstanceEntity, error)
	// DeleteByBusinessKey 根据业务键删除
	DeleteByBusinessKey(projectID string, apiType entity.ApiType, businessID string) (int64, error)
	// BatchDeleteByBusinessKeys 批量删除
	BatchDeleteByBusinessKeys(projectID string, apiType entity.ApiType, businessIDs []string) (int64, error)
}
