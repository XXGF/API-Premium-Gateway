package service

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
)

// AffinityService 亲和性服务
type AffinityService struct {
	bindingCache *cache.Cache
}

// NewAffinityService 创建亲和性服务
func NewAffinityService() *AffinityService {
	// 默认过期时间60分钟，清理间隔10分钟
	c := cache.New(60*time.Minute, 10*time.Minute)
	return &AffinityService{bindingCache: c}
}

// GetBoundInstance 获取绑定的实例ID
func (s *AffinityService) GetBoundInstance(affinityType, affinityKey string) string {
	bindingKey := buildBindingKey(affinityType, affinityKey)
	val, found := s.bindingCache.Get(bindingKey)
	if !found {
		log.Debug().Str("key", bindingKey).Msg("未找到亲和性绑定")
		return ""
	}

	binding, ok := val.(*entity.AffinityBinding)
	if !ok {
		return ""
	}

	if binding.IsExpired() {
		s.bindingCache.Delete(bindingKey)
		log.Debug().Str("key", bindingKey).Msg("亲和性绑定已过期，已清除")
		return ""
	}

	log.Debug().Str("key", bindingKey).Str("instanceId", binding.InstanceID).Int("useCount", binding.UseCount).Msg("找到亲和性绑定")
	return binding.InstanceID
}

// CreateBinding 创建新的绑定
func (s *AffinityService) CreateBinding(affinityType, affinityKey, instanceID string) {
	bindingKey := buildBindingKey(affinityType, affinityKey)
	binding := entity.NewAffinityBinding(instanceID)
	s.bindingCache.Set(bindingKey, binding, 60*time.Minute)
	log.Info().Str("key", bindingKey).Str("instanceId", instanceID).Msg("创建亲和性绑定")
}

// RefreshBinding 刷新绑定
func (s *AffinityService) RefreshBinding(affinityType, affinityKey, instanceID string) {
	bindingKey := buildBindingKey(affinityType, affinityKey)
	val, found := s.bindingCache.Get(bindingKey)
	if !found {
		return
	}

	binding, ok := val.(*entity.AffinityBinding)
	if !ok || binding.InstanceID != instanceID {
		log.Warn().Str("key", bindingKey).Str("expected", instanceID).Msg("尝试刷新不匹配的亲和性绑定")
		return
	}

	newExpireTime := time.Now().Add(60 * time.Minute)
	refreshed := binding.WithNewExpireTime(newExpireTime)
	s.bindingCache.Set(bindingKey, refreshed, 60*time.Minute)

	log.Debug().Str("key", bindingKey).Str("instanceId", instanceID).Int("useCount", refreshed.UseCount).Msg("刷新亲和性绑定")
}

// ClearBinding 清除绑定
func (s *AffinityService) ClearBinding(affinityType, affinityKey string) {
	bindingKey := buildBindingKey(affinityType, affinityKey)
	s.bindingCache.Delete(bindingKey)
	log.Info().Str("key", bindingKey).Msg("清除亲和性绑定")
}

// GetBindingCount 获取当前绑定数量
func (s *AffinityService) GetBindingCount() int {
	return s.bindingCache.ItemCount()
}

// ClearAllBindings 清除所有绑定
func (s *AffinityService) ClearAllBindings() {
	s.bindingCache.Flush()
	log.Info().Msg("清除所有亲和性绑定")
}

func buildBindingKey(affinityType, affinityKey string) string {
	return affinityType + ":" + affinityKey
}
