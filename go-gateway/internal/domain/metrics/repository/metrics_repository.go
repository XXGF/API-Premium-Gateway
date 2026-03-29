package repository

import (
	"time"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// MetricsRepository 指标仓储接口
type MetricsRepository interface {
	// Insert 插入指标记录
	Insert(metrics *entity.InstanceMetricsEntity) error
	// UpdateByID 根据ID更新指标记录
	UpdateByID(metrics *entity.InstanceMetricsEntity) error
	// SelectByRegistryIDAndWindow 根据实例ID和时间窗口查询
	SelectByRegistryIDAndWindow(registryID string, timestampWindow time.Time) (*entity.InstanceMetricsEntity, error)
	// SelectByRegistryIDsAfterTime 根据实例ID列表和时间查询
	SelectByRegistryIDsAfterTime(registryIDs []string, cutoffTime time.Time) ([]*entity.InstanceMetricsEntity, error)
}
