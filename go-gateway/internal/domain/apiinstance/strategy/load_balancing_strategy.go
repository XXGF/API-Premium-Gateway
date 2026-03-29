package strategy

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// LoadBalancingStrategy 负载均衡策略接口
type LoadBalancingStrategy interface {
	// GetStrategyName 策略名称
	GetStrategyName() string
	// GetDescription 策略描述
	GetDescription() string
	// GetStrategyType 获取策略类型
	GetStrategyType() entity.LoadBalancingType
	// SelectInstance 选择最佳实例
	SelectInstance(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) (*entity.ApiInstanceEntity, error)
}
