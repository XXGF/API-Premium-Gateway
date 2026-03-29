package service

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/strategy"
)

// AffinityAwareStrategyDecorator 亲和性感知的策略装饰器。
//
// 采用装饰器模式（Decorator Pattern），在负载均衡策略之上叠加亲和性绑定逻辑。
// 工作流程：
//  1. 检查是否有亲和性要求，没有则直接使用负载均衡策略
//  2. 查找现有绑定，如果绑定实例可用则直接返回
//  3. 绑定实例不可用时，根据亲和性强度决定行为（报错/重新选择）
//  4. 没有绑定时，使用策略选择新实例并创建绑定
//
// 注意：当前版本亲和性功能暂时禁用（与 Java 版本一致）。
type AffinityAwareStrategyDecorator struct {
	affinityService *AffinityService
}

// NewAffinityAwareStrategyDecorator 创建亲和性装饰器
func NewAffinityAwareStrategyDecorator(affinityService *AffinityService) *AffinityAwareStrategyDecorator {
	return &AffinityAwareStrategyDecorator{affinityService: affinityService}
}

// SelectInstanceWithAffinity 带亲和性的实例选择。
//
// 在负载均衡策略之上叠加亲和性绑定逻辑：
//   - 无亲和性要求：直接使用负载均衡策略
//   - 有绑定且可用：刷新绑定并返回绑定实例
//   - 有绑定但不可用：根据强度处理（strict 报错 / preferred 重选）
//   - 无绑定：策略选择新实例并创建绑定
func (d *AffinityAwareStrategyDecorator) SelectInstanceWithAffinity(
	candidates []*entity.ApiInstanceEntity,
	metricsMap map[string]*metricsEntity.InstanceMetricsEntity,
	strat strategy.LoadBalancingStrategy,
	affinityContext *entity.AffinityContext,
) (*entity.ApiInstanceEntity, error) {

	// 当前版本亲和性功能暂时禁用（与 Java 版本一致）
	affinityContext = nil

	// 1. 如果没有亲和性要求，直接使用负载均衡策略
	if affinityContext == nil || !affinityContext.IsValid() {
		log.Debug().Str("strategy", strat.GetStrategyName()).Msg("无亲和性要求，直接使用负载均衡策略")
		return strat.SelectInstance(candidates, metricsMap)
	}

	// 2. 检查是否有现有的亲和性绑定
	boundInstanceID := d.affinityService.GetBoundInstance(affinityContext.AffinityType, affinityContext.AffinityKey)

	if boundInstanceID != "" {
		// 3. 查找绑定的实例是否在候选列表中且健康
		boundInstance := findInstanceByID(candidates, boundInstanceID)

		if boundInstance != nil {
			d.affinityService.RefreshBinding(affinityContext.AffinityType, affinityContext.AffinityKey, boundInstanceID)
			log.Debug().Str("bindingKey", affinityContext.GetBindingKey()).Str("instanceId", boundInstanceID).Msg("使用亲和性绑定实例")
			return boundInstance, nil
		}

		// 绑定的实例不可用
		log.Warn().Str("bindingKey", affinityContext.GetBindingKey()).Str("instanceId", boundInstanceID).Msg("亲和性绑定的实例不可用")
		return d.handleUnavailableBinding(candidates, metricsMap, strat, affinityContext)
	}

	// 4. 没有现有绑定，使用负载均衡策略选择新实例并创建绑定
	selectedInstance, err := strat.SelectInstance(candidates, metricsMap)
	if err != nil {
		return nil, err
	}

	if selectedInstance != nil {
		d.affinityService.CreateBinding(affinityContext.AffinityType, affinityContext.AffinityKey, selectedInstance.ID)
		log.Info().Str("bindingKey", affinityContext.GetBindingKey()).Str("instanceId", selectedInstance.ID).Msg("创建新的亲和性绑定")
	}

	return selectedInstance, nil
}

// handleUnavailableBinding 处理绑定实例不可用的情况。
//
// 根据亲和性强度采取不同策略：
//   - strict：直接返回错误，不允许切换到其他实例
//   - preferred：清除旧绑定，使用策略重新选择并创建新绑定
//   - 其他：直接使用负载均衡策略
func (d *AffinityAwareStrategyDecorator) handleUnavailableBinding(
	candidates []*entity.ApiInstanceEntity,
	metricsMap map[string]*metricsEntity.InstanceMetricsEntity,
	strat strategy.LoadBalancingStrategy,
	affinityContext *entity.AffinityContext,
) (*entity.ApiInstanceEntity, error) {

	switch affinityContext.Strength {
	case entity.AffinityStrengthStrict:
		errMsg := fmt.Sprintf("严格亲和性模式下，绑定实例不可用: %s", affinityContext.GetBindingKey())
		log.Error().Msg(errMsg)
		return nil, fmt.Errorf(errMsg)

	case entity.AffinityStrengthPreferred:
		log.Info().Str("bindingKey", affinityContext.GetBindingKey()).Msg("优先亲和性模式下，清除不可用绑定并重新选择")
		d.affinityService.ClearBinding(affinityContext.AffinityType, affinityContext.AffinityKey)

		newInstance, err := strat.SelectInstance(candidates, metricsMap)
		if err != nil {
			return nil, err
		}
		if newInstance != nil {
			d.affinityService.CreateBinding(affinityContext.AffinityType, affinityContext.AffinityKey, newInstance.ID)
			log.Info().Str("bindingKey", affinityContext.GetBindingKey()).Str("instanceId", newInstance.ID).Msg("重新创建亲和性绑定")
		}
		return newInstance, nil

	default:
		log.Debug().Msg("无亲和性模式，直接使用负载均衡策略")
		return strat.SelectInstance(candidates, metricsMap)
	}
}

// findInstanceByID 根据实例ID查找实例
func findInstanceByID(candidates []*entity.ApiInstanceEntity, instanceID string) *entity.ApiInstanceEntity {
	for _, instance := range candidates {
		if instance.ID == instanceID {
			return instance
		}
	}
	return nil
}
