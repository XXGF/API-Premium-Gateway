package service

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/command"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/repository"
)

// 选择算法配置常量
const (
	CircuitBreakerErrorRateThreshold = 0.5
	CircuitBreakerMinRequestCount    = 10
	ShortTermWindowMinutes           = 5
	LatencyScoreMaxMs                = 5000
)

// MetricsCollectionDomainService 指标收集领域服务
type MetricsCollectionDomainService struct {
	metricsRepository repository.MetricsRepository
	mu                sync.Mutex
}

// NewMetricsCollectionDomainService 创建指标收集领域服务
func NewMetricsCollectionDomainService(repo repository.MetricsRepository) *MetricsCollectionDomainService {
	return &MetricsCollectionDomainService{
		metricsRepository: repo,
	}
}

// RecordCallResult 记录API调用结果
func (s *MetricsCollectionDomainService) RecordCallResult(cmd *command.CallResultCommand) error {
	log.Info().Str("instanceId", cmd.InstanceID).Bool("success", cmd.Success).Msg("开始记录调用结果")

	// 获取或创建当前时间窗口的指标记录
	currentWindow := getCurrentTimeWindow()
	metrics, err := s.getOrCreateMetrics(cmd.InstanceID, currentWindow)
	if err != nil {
		return err
	}

	// 更新指标
	s.updateMetrics(metrics, cmd.Success, cmd.LatencyMs, cmd.UsageMetrics)

	// 更新Gateway状态
	s.updateGatewayStatus(metrics)

	// 保存指标
	if metrics.ID == "" {
		if err := s.metricsRepository.Insert(metrics); err != nil {
			return err
		}
		log.Debug().Str("instanceId", cmd.InstanceID).Msg("创建新的指标记录")
	} else {
		if err := s.metricsRepository.UpdateByID(metrics); err != nil {
			return err
		}
		log.Debug().Str("instanceId", cmd.InstanceID).Msg("更新指标记录")
	}

	log.Info().Str("instanceId", cmd.InstanceID).Msg("调用结果记录完成")
	return nil
}

// GetInstanceMetrics 获取实例指标数据
func (s *MetricsCollectionDomainService) GetInstanceMetrics(instanceIDs []string) (map[string]*entity.InstanceMetricsEntity, error) {
	if len(instanceIDs) == 0 {
		return make(map[string]*entity.InstanceMetricsEntity), nil
	}

	cutoffTime := time.Now().Add(-time.Duration(ShortTermWindowMinutes) * time.Minute)
	metricsList, err := s.metricsRepository.SelectByRegistryIDsAfterTime(instanceIDs, cutoffTime)
	if err != nil {
		return nil, err
	}

	log.Debug().Int("count", len(metricsList)).Int("instanceCount", len(instanceIDs)).Msg("查询到指标数据")

	// 聚合同一实例的指标数据（取最新的）
	result := make(map[string]*entity.InstanceMetricsEntity)
	for _, metrics := range metricsList {
		existing, ok := result[metrics.RegistryID]
		if !ok || metrics.TimestampWindow.After(existing.TimestampWindow) {
			result[metrics.RegistryID] = metrics
		}
	}

	return result, nil
}

// getCurrentTimeWindow 获取当前时间窗口（按分钟截断）
func getCurrentTimeWindow() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
}

// getOrCreateMetrics 获取或创建指标记录
func (s *MetricsCollectionDomainService) getOrCreateMetrics(instanceID string, timeWindow time.Time) (*entity.InstanceMetricsEntity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics, err := s.metricsRepository.SelectByRegistryIDAndWindow(instanceID, timeWindow)
	if err != nil {
		return nil, err
	}

	if metrics == nil {
		metrics = entity.NewInstanceMetricsEntity(instanceID, timeWindow)
	} else {
		metrics.LastReportedAt = time.Now()
	}

	return metrics, nil
}

// updateMetrics 更新指标数据
func (s *MetricsCollectionDomainService) updateMetrics(metrics *entity.InstanceMetricsEntity, success bool, latencyMs int64, usageMetrics map[string]interface{}) {
	if success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}

	metrics.TotalLatencyMs += latencyMs

	// 合并使用指标
	if len(usageMetrics) > 0 {
		if metrics.AdditionalMetrics == nil {
			metrics.AdditionalMetrics = make(entity.JSONBMap)
		}
		for k, v := range usageMetrics {
			metrics.AdditionalMetrics[k] = v
		}
	}
}

// updateGatewayStatus 更新Gateway状态
func (s *MetricsCollectionDomainService) updateGatewayStatus(metrics *entity.InstanceMetricsEntity) {
	successRate := metrics.GetSuccessRate()
	totalCalls := metrics.GetTotalCount()

	// 如果调用次数太少，保持健康状态
	if totalCalls < CircuitBreakerMinRequestCount {
		metrics.UpdateGatewayStatus(entity.GatewayStatusHealthy)
		return
	}

	// 判断是否需要熔断
	if successRate < CircuitBreakerErrorRateThreshold {
		log.Warn().Str("instanceId", metrics.RegistryID).Float64("successRate", successRate).Int64("totalCalls", totalCalls).Msg("实例错误率过高，触发熔断")
		metrics.UpdateGatewayStatus(entity.GatewayStatusCircuitBreakerOpen)
		return
	}

	// 判断是否降级
	avgLatency := metrics.GetAverageLatency()
	if avgLatency > LatencyScoreMaxMs {
		log.Warn().Str("instanceId", metrics.RegistryID).Float64("avgLatency", avgLatency).Msg("实例延迟过高，标记为降级")
		metrics.UpdateGatewayStatus(entity.GatewayStatusDegraded)
		return
	}

	// 正常健康状态
	metrics.UpdateGatewayStatus(entity.GatewayStatusHealthy)
}
