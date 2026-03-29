package persistence

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// MetricsRepositoryImpl 指标仓储实现
type MetricsRepositoryImpl struct {
	db *gorm.DB
}

// NewMetricsRepositoryImpl 创建指标仓储实现
func NewMetricsRepositoryImpl(db *gorm.DB) *MetricsRepositoryImpl {
	return &MetricsRepositoryImpl{db: db}
}

func (r *MetricsRepositoryImpl) Insert(metrics *entity.InstanceMetricsEntity) error {
	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}
	return r.db.Create(metrics).Error
}

func (r *MetricsRepositoryImpl) UpdateByID(metrics *entity.InstanceMetricsEntity) error {
	return r.db.Save(metrics).Error
}

func (r *MetricsRepositoryImpl) SelectByRegistryIDAndWindow(registryID string, timestampWindow time.Time) (*entity.InstanceMetricsEntity, error) {
	var metrics entity.InstanceMetricsEntity
	result := r.db.Where("registry_id = ? AND timestamp_window = ?", registryID, timestampWindow).First(&metrics)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &metrics, nil
}

func (r *MetricsRepositoryImpl) SelectByRegistryIDsAfterTime(registryIDs []string, cutoffTime time.Time) ([]*entity.InstanceMetricsEntity, error) {
	var metricsList []*entity.InstanceMetricsEntity
	result := r.db.Where("registry_id IN ? AND timestamp_window >= ?", registryIDs, cutoffTime).
		Order("timestamp_window DESC").
		Find(&metricsList)
	return metricsList, result.Error
}
