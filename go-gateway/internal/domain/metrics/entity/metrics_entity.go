// Package entity 定义了 Metrics 领域的核心实体和值对象。
//
// 该包属于 DDD 架构的领域层，包含 InstanceMetricsEntity 聚合根，
// 用于记录 API 实例在特定时间窗口内的调用指标（成功率、延迟、并发等），
// 以及 GatewayStatus 枚举和 JSONBMap 类型映射。
package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"database/sql/driver"
)

// GatewayStatus 定义了 Gateway 对 API 实例的健康状态评估结果。
//
// 该状态由 MetricsCollectionDomainService 根据调用指标自动计算，
// 决定了实例是否可以参与路由选择。
type GatewayStatus string

const (
	GatewayStatusHealthy           GatewayStatus = "HEALTHY"             // 健康状态，可正常参与路由
	GatewayStatusDegraded          GatewayStatus = "DEGRADED"            // 降级状态，平均延迟超过阈值（5000ms）
	GatewayStatusFaulty            GatewayStatus = "FAULTY"              // 故障状态
	GatewayStatusCircuitBreakerOpen GatewayStatus = "CIRCUIT_BREAKER_OPEN" // 熔断状态，成功率低于50%且调用次数≥10
)

// GatewayStatusFromCode 根据代码获取状态
func GatewayStatusFromCode(code string) (GatewayStatus, error) {
	switch code {
	case "HEALTHY":
		return GatewayStatusHealthy, nil
	case "DEGRADED":
		return GatewayStatusDegraded, nil
	case "FAULTY":
		return GatewayStatusFaulty, nil
	case "CIRCUIT_BREAKER_OPEN":
		return GatewayStatusCircuitBreakerOpen, nil
	default:
		return "", fmt.Errorf("未知的网关状态代码: %s", code)
	}
}

// IsAvailableForRouting 判断是否可用于路由
func (s GatewayStatus) IsAvailableForRouting() bool {
	return s == GatewayStatusHealthy || s == GatewayStatusDegraded
}

// JSONBMap 是 PostgreSQL JSONB 类型的 Go 映射。
//
// 实现了 driver.Valuer 和 sql.Scanner 接口，
// 使 GORM 能够自动序列化/反序列化 JSONB 字段。
type JSONBMap map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONBMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONBMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONBMap)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("无法将 %T 转换为 JSONBMap", value)
	}
	return json.Unmarshal(bytes, j)
}

// InstanceMetricsEntity 是 Metrics 领域的聚合根。
//
// 记录某个 API 实例在特定时间窗口（分钟级）内的调用指标，
// 包括成功/失败次数、总延迟、并发数和健康状态。
// 对应数据库表 api_instance_metrics。
type InstanceMetricsEntity struct {
	ID                   string        `gorm:"column:id;primaryKey" json:"id"`                                  // 唯一标识符（UUID）
	RegistryID           string        `gorm:"column:registry_id" json:"registry_id"`                           // 关联的 API 实例 ID
	TimestampWindow      time.Time     `gorm:"column:timestamp_window" json:"timestamp_window"`                 // 时间窗口（按分钟截断）
	SuccessCount         int64         `gorm:"column:success_count" json:"success_count"`                       // 窗口内成功调用次数
	FailureCount         int64         `gorm:"column:failure_count" json:"failure_count"`                       // 窗口内失败调用次数
	TotalLatencyMs       int64         `gorm:"column:total_latency_ms" json:"total_latency_ms"`                 // 窗口内总延迟（毫秒）
	Concurrency          int           `gorm:"column:concurrency" json:"concurrency"`                           // 当前并发数
	CurrentGatewayStatus GatewayStatus `gorm:"column:current_gateway_status" json:"current_gateway_status"`     // Gateway 评估的健康状态
	LastReportedAt       time.Time     `gorm:"column:last_reported_at" json:"last_reported_at"`                 // 最后上报时间
	AdditionalMetrics    JSONBMap      `gorm:"column:additional_metrics;type:jsonb" json:"additional_metrics"` // 额外指标（如 token 消耗）
}

// TableName 指定表名
func (InstanceMetricsEntity) TableName() string {
	return "api_instance_metrics"
}

// NewInstanceMetricsEntity 创建指标实体
func NewInstanceMetricsEntity(registryID string, timestampWindow time.Time) *InstanceMetricsEntity {
	return &InstanceMetricsEntity{
		RegistryID:           registryID,
		TimestampWindow:      timestampWindow,
		SuccessCount:         0,
		FailureCount:         0,
		TotalLatencyMs:       0,
		Concurrency:          0,
		CurrentGatewayStatus: GatewayStatusHealthy,
		LastReportedAt:       time.Now(),
	}
}

// GetSuccessRate 计算当前时间窗口内的调用成功率。
//
// 返回值范围 [0.0, 1.0]。如果没有任何调用记录，返回 1.0（乐观默认值）。
func (m *InstanceMetricsEntity) GetSuccessRate() float64 {
	total := m.SuccessCount + m.FailureCount
	if total == 0 {
		return 1.0
	}
	return float64(m.SuccessCount) / float64(total)
}

// GetAverageLatency 计算当前时间窗口内的平均调用延迟（毫秒）。
//
// 如果没有任何调用记录，返回 0.0。
func (m *InstanceMetricsEntity) GetAverageLatency() float64 {
	total := m.SuccessCount + m.FailureCount
	if total == 0 {
		return 0.0
	}
	return float64(m.TotalLatencyMs) / float64(total)
}

// GetTotalCount 获取总调用次数
func (m *InstanceMetricsEntity) GetTotalCount() int64 {
	return m.SuccessCount + m.FailureCount
}

// IsHealthy 检查是否健康
func (m *InstanceMetricsEntity) IsHealthy() bool {
	return m.CurrentGatewayStatus == GatewayStatusHealthy
}

// IsCircuitBreakerOpen 检查是否被熔断
func (m *InstanceMetricsEntity) IsCircuitBreakerOpen() bool {
	return m.CurrentGatewayStatus == GatewayStatusCircuitBreakerOpen
}

// UpdateGatewayStatus 更新网关状态
func (m *InstanceMetricsEntity) UpdateGatewayStatus(status GatewayStatus) {
	m.CurrentGatewayStatus = status
}
