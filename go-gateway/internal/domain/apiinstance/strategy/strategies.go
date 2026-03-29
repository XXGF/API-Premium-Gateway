// Package strategy 实现了 API 实例的负载均衡策略。
//
// 该包采用策略模式（Strategy Pattern）设计，通过 LoadBalancingStrategy 接口
// 统一抽象了 4 种负载均衡算法，并由 LoadBalancingStrategyFactory 工厂管理。
//
// 支持的策略：
//   - RoundRobinStrategy：轮询策略，依次选择每个可用实例
//   - LatencyFirstStrategy：延迟优先，选择平均延迟最低的实例
//   - SuccessRateFirstStrategy：成功率优先，选择历史成功率最高的实例
//   - SmartStrategy：智能策略，综合评分（成功率×0.4 + 延迟×0.4 + 负载×0.2）
package strategy

import (
	"fmt"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// RoundRobinStrategy 轮询负载均衡策略。
//
// 通过原子计数器实现线程安全的轮询选择，依次选择每个可用实例。
// 适用于实例性能相近的场景。
type RoundRobinStrategy struct {
	counter uint64 // 原子计数器，用于轮询索引
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

// LatencyFirstStrategy 延迟优先负载均衡策略。
//
// 遍历所有候选实例，选择平均延迟最低的实例。
// 对于没有指标数据的新实例，使用冷启动默认延迟 1000ms。
// 适用于对响应时间敏感的场景。
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

// SuccessRateFirstStrategy 成功率优先负载均衡策略。
//
// 遍历所有候选实例，选择历史成功率最高的实例。
// 对于没有指标数据的新实例，使用冷启动默认成功率 1.0。
// 适用于对稳定性要求高的场景。
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

// SmartStrategy 智能负载均衡策略（默认策略）。
//
// 通过综合评分算法选择最优实例，考虑成功率、延迟、负载等多个指标。
//
// 综合得分公式：
//
//	综合得分 = 成功率得分 × 0.4 + 延迟得分 × 0.4 + 负载得分 × 0.2
//
// 其中：
//   - 成功率得分 = successRate × 100
//   - 延迟得分 = 100 × (2000 / (2000 + avgLatency))
//   - 负载得分 = 100 × (100 / (100 + concurrency))
type SmartStrategy struct{}

// NewSmartStrategy 创建智能策略
func NewSmartStrategy() *SmartStrategy {
	return &SmartStrategy{}
}

// 智能策略的评分权重和参数常量
const (
	successRateWeight    = 0.4    // 成功率权重，占总分的40%
	latencyWeight        = 0.4    // 延迟权重，占总分的40%
	loadWeight           = 0.2    // 负载权重，占总分的20%
	maxScore             = 100.0  // 单项最高得分
	latencyScoreMaxMs    = 2000.0 // 延迟评分的参考基准值（ms）
	loadScoreMaxConc     = 100    // 负载评分的参考基准并发数
	coldStartSuccessRate = 1.0    // 冷启动默认成功率（乐观值）
	coldStartLatency     = 1000.0 // 冷启动默认延迟（ms）
	coldStartConcurrency = 1      // 冷启动默认并发数
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

// calculateComprehensiveScore 计算实例的综合得分。
//
// 对于没有指标数据的新实例（冷启动），使用乐观默认值：
//   - 成功率 = 1.0，延迟 = 1000ms，并发 = 1
//
// 这确保新实例有机会被选中，从而获得真实的指标数据。
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

// calculateLatencyScore 计算延迟得分。
//
// 使用反比例函数：score = 100 × (2000 / (2000 + latency))
// 延迟越低得分越高，延迟为 0 时得分为 100，延迟为 2000ms 时得分为 50。
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

// calculateLoadScore 计算负载得分。
//
// 使用反比例函数：score = 100 × (100 / (100 + concurrency))
// 并发越低得分越高，并发为 0 时得分为 100，并发为 100 时得分为 50。
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

// LoadBalancingStrategyFactory 负载均衡策略工厂。
//
// 负责管理所有已注册的负载均衡策略实例，
// 提供按类型获取策略和获取默认策略的能力。
// 在应用启动时初始化，自动注册所有 4 种策略实现。
type LoadBalancingStrategyFactory struct {
	strategies map[entity.LoadBalancingType]LoadBalancingStrategy // 策略注册表
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

// filterNonCircuitBroken 过滤掉被熔断的实例。
//
// 遍历候选实例列表，排除指标状态为 CIRCUIT_BREAKER_OPEN 的实例。
// 对于没有指标数据的实例（新实例），默认保留。
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

// getAverageLatency 获取实例平均延迟。
//
// 如果实例没有指标数据（冷启动），返回默认值 1000ms。
func getAverageLatency(instanceID string, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) float64 {
	metrics, ok := metricsMap[instanceID]
	if !ok || metrics == nil {
		return coldStartLatency
	}
	return metrics.GetAverageLatency()
}

// getSuccessRate 获取实例成功率。
//
// 如果实例没有指标数据（冷启动），返回默认值 1.0。
func getSuccessRate(instanceID string, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) float64 {
	metrics, ok := metricsMap[instanceID]
	if !ok || metrics == nil {
		return coldStartSuccessRate
	}
	return metrics.GetSuccessRate()
}
