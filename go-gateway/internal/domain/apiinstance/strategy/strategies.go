package strategy

import (
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// RoundRobinStrategy 轮询负载均衡策略
type RoundRobinStrategy struct {
	counter uint64
}

// NewRoundRobinStrategy 创建轮询策略
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

func (s *RoundRobinStrategy) GetStrategyName() string {
	return "ROUND_ROBIN"
}

func (s *RoundRobinStrategy) GetDescription() string {
	return "轮询策略：依次选择每个可用实例，适用于实例性能相近的场景"
}

func (s *RoundRobinStrategy) GetStrategyType() entity.LoadBalancingType {
	return entity.LoadBalancingTypeRoundRobin
}

func (s *RoundRobinStrategy) SelectInstance(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) (*entity.ApiInstanceEntity, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("候选实例列表不能为空")
	}

	// 过滤掉被熔断的实例
	available := filterNonCircuitBroken(candidates, metricsMap)
	if len(available) == 0 {
		log.Warn().Msg("所有实例都被熔断，返回第一个实例")
		return candidates[0], nil
	}

	// 轮询选择
	index := atomic.AddUint64(&s.counter, 1) % uint64(len(available))
	selected := available[index]

	log.Debug().Str("businessId", selected.BusinessID).Uint64("counter", s.counter).Msg("轮询策略选择实例")
	return selected, nil
}

// LatencyFirstStrategy 延迟优先负载均衡策略
type LatencyFirstStrategy struct{}

// NewLatencyFirstStrategy 创建延迟优先策略
func NewLatencyFirstStrategy() *LatencyFirstStrategy {
	return &LatencyFirstStrategy{}
}

func (s *LatencyFirstStrategy) GetStrategyName() string {
	return "LATENCY_FIRST"
}

func (s *LatencyFirstStrategy) GetDescription() string {
	return "延迟优先策略：选择平均延迟最低的实例，适用于对响应时间敏感的场景"
}

func (s *LatencyFirstStrategy) GetStrategyType() entity.LoadBalancingType {
	return entity.LoadBalancingTypeLatencyFirst
}

func (s *LatencyFirstStrategy) SelectInstance(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) (*entity.ApiInstanceEntity, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("候选实例列表不能为空")
	}

	available := filterNonCircuitBroken(candidates, metricsMap)
	if len(available) == 0 {
		log.Warn().Msg("所有实例都被熔断，返回第一个实例")
		return candidates[0], nil
	}

	// 按延迟排序，选择延迟最低的实例
	var selected *entity.ApiInstanceEntity
	minLatency := float64(1<<63 - 1)

	for _, instance := range available {
		latency := getAverageLatency(instance.ID, metricsMap)
		if latency < minLatency {
			minLatency = latency
			selected = instance
		}
	}

	if selected == nil {
		selected = available[0]
	}

	log.Debug().Str("businessId", selected.BusinessID).Float64("latency", minLatency).Msg("延迟优先策略选择实例")
	return selected, nil
}

// SuccessRateFirstStrategy 成功率优先负载均衡策略
type SuccessRateFirstStrategy struct{}

// NewSuccessRateFirstStrategy 创建成功率优先策略
func NewSuccessRateFirstStrategy() *SuccessRateFirstStrategy {
	return &SuccessRateFirstStrategy{}
}

func (s *SuccessRateFirstStrategy) GetStrategyName() string {
	return "SUCCESS_RATE_FIRST"
}

func (s *SuccessRateFirstStrategy) GetDescription() string {
	return "成功率优先策略：选择历史成功率最高的实例，适用于对稳定性要求高的场景"
}

func (s *SuccessRateFirstStrategy) GetStrategyType() entity.LoadBalancingType {
	return entity.LoadBalancingTypeSuccessRateFirst
}

func (s *SuccessRateFirstStrategy) SelectInstance(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) (*entity.ApiInstanceEntity, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("候选实例列表不能为空")
	}

	available := filterNonCircuitBroken(candidates, metricsMap)
	if len(available) == 0 {
		log.Warn().Msg("所有实例都被熔断，返回第一个实例")
		return candidates[0], nil
	}

	// 按成功率排序，选择成功率最高的实例
	var selected *entity.ApiInstanceEntity
	maxSuccessRate := -1.0

	for _, instance := range available {
		successRate := getSuccessRate(instance.ID, metricsMap)
		if successRate > maxSuccessRate {
			maxSuccessRate = successRate
			selected = instance
		}
	}

	if selected == nil {
		selected = available[0]
	}

	log.Debug().Str("businessId", selected.BusinessID).Float64("successRate", maxSuccessRate).Msg("成功率优先策略选择实例")
	return selected, nil
}

// SmartStrategy 智能负载均衡策略
type SmartStrategy struct{}

// NewSmartStrategy 创建智能策略
func NewSmartStrategy() *SmartStrategy {
	return &SmartStrategy{}
}

const (
	successRateWeight    = 0.4
	latencyWeight        = 0.4
	loadWeight           = 0.2
	maxScore             = 100.0
	latencyScoreMaxMs    = 2000.0
	loadScoreMaxConc     = 100
	coldStartSuccessRate = 1.0
	coldStartLatency     = 1000.0
	coldStartConcurrency = 1
)

func (s *SmartStrategy) GetStrategyName() string {
	return "智能策略"
}

func (s *SmartStrategy) GetDescription() string {
	return "通过综合评分算法选择最优实例，考虑成功率、延迟、负载等多个指标"
}

func (s *SmartStrategy) GetStrategyType() entity.LoadBalancingType {
	return entity.LoadBalancingTypeSmart
}

func (s *SmartStrategy) SelectInstance(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) (*entity.ApiInstanceEntity, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("候选实例列表不能为空")
	}

	log.Debug().Int("count", len(candidates)).Msg("智能策略开始综合评分")

	available := filterNonCircuitBroken(candidates, metricsMap)
	if len(available) == 0 {
		log.Warn().Msg("所有实例都被熔断，返回第一个实例")
		return candidates[0], nil
	}

	// 计算每个实例的综合得分，选择得分最高的实例
	var selected *entity.ApiInstanceEntity
	maxScoreVal := -1.0

	for _, instance := range available {
		score := s.calculateComprehensiveScore(instance, metricsMap)
		if score > maxScoreVal {
			maxScoreVal = score
			selected = instance
		}
	}

	if selected == nil {
		selected = available[0]
	}

	log.Info().Str("businessId", selected.BusinessID).Float64("score", maxScoreVal).Msg("智能策略选择实例")
	return selected, nil
}

// calculateComprehensiveScore 计算实例的综合得分
func (s *SmartStrategy) calculateComprehensiveScore(instance *entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) float64 {
	metrics := metricsMap[instance.ID]

	successRate := coldStartSuccessRate
	latency := coldStartLatency
	concurrency := coldStartConcurrency

	if metrics != nil {
		successRate = metrics.GetSuccessRate()
		latency = metrics.GetAverageLatency()
		concurrency = metrics.Concurrency
	}

	// 计算各项得分
	successRateScore := successRate * maxScore
	latencyScore := s.calculateLatencyScore(latency)
	loadScore := s.calculateLoadScore(concurrency)

	// 计算综合得分
	return successRateScore*successRateWeight + latencyScore*latencyWeight + loadScore*loadWeight
}

// calculateLatencyScore 计算延迟得分
func (s *SmartStrategy) calculateLatencyScore(latency float64) float64 {
	if latency <= 0 {
		return maxScore
	}
	score := maxScore * (latencyScoreMaxMs / (latencyScoreMaxMs + latency))
	if score < 0 {
		return 0
	}
	return score
}

// calculateLoadScore 计算负载得分
func (s *SmartStrategy) calculateLoadScore(concurrency int) float64 {
	if concurrency <= 0 {
		return maxScore
	}
	score := maxScore * (float64(loadScoreMaxConc) / (float64(loadScoreMaxConc) + float64(concurrency)))
	if score < 0 {
		return 0
	}
	return score
}

// LoadBalancingStrategyFactory 负载均衡策略工厂
type LoadBalancingStrategyFactory struct {
	strategies map[entity.LoadBalancingType]LoadBalancingStrategy
}

// NewLoadBalancingStrategyFactory 创建策略工厂
func NewLoadBalancingStrategyFactory() *LoadBalancingStrategyFactory {
	factory := &LoadBalancingStrategyFactory{
		strategies: make(map[entity.LoadBalancingType]LoadBalancingStrategy),
	}

	// 注册所有策略
	roundRobin := NewRoundRobinStrategy()
	latencyFirst := NewLatencyFirstStrategy()
	successRateFirst := NewSuccessRateFirstStrategy()
	smart := NewSmartStrategy()

	factory.strategies[roundRobin.GetStrategyType()] = roundRobin
	factory.strategies[latencyFirst.GetStrategyType()] = latencyFirst
	factory.strategies[successRateFirst.GetStrategyType()] = successRateFirst
	factory.strategies[smart.GetStrategyType()] = smart

	log.Info().Int("count", len(factory.strategies)).Msg("负载均衡策略初始化完成")
	return factory
}

// GetStrategy 根据类型获取策略
func (f *LoadBalancingStrategyFactory) GetStrategy(lbType entity.LoadBalancingType) (LoadBalancingStrategy, error) {
	strategy, ok := f.strategies[lbType]
	if !ok {
		return nil, fmt.Errorf("未找到负载均衡策略类型 %s 的实现", lbType)
	}
	return strategy, nil
}

// GetDefaultStrategy 获取默认策略（智能策略）
func (f *LoadBalancingStrategyFactory) GetDefaultStrategy() LoadBalancingStrategy {
	strategy, _ := f.GetStrategy(entity.LoadBalancingTypeSmart)
	return strategy
}

// ===== 辅助函数 =====

// filterNonCircuitBroken 过滤掉被熔断的实例
func filterNonCircuitBroken(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) []*entity.ApiInstanceEntity {
	var result []*entity.ApiInstanceEntity
	for _, instance := range candidates {
		metrics, ok := metricsMap[instance.ID]
		if !ok || !metrics.IsCircuitBreakerOpen() {
			result = append(result, instance)
		}
	}
	return result
}

// getAverageLatency 获取实例平均延迟
func getAverageLatency(instanceID string, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) float64 {
	metrics, ok := metricsMap[instanceID]
	if !ok || metrics == nil {
		return coldStartLatency
	}
	return metrics.GetAverageLatency()
}

// getSuccessRate 获取实例成功率
func getSuccessRate(instanceID string, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) float64 {
	metrics, ok := metricsMap[instanceID]
	if !ok || metrics == nil {
		return coldStartSuccessRate
	}
	return metrics.GetSuccessRate()
}
