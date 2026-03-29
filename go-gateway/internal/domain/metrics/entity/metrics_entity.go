package entity

import (
	"encoding/json"
	"fmt"
	"time"

	"database/sql/driver"
)

// GatewayStatus Gateway 判断的 API 实例状态
type GatewayStatus string

const (
	GatewayStatusHealthy           GatewayStatus = "HEALTHY"
	GatewayStatusDegraded          GatewayStatus = "DEGRADED"
	GatewayStatusFaulty            GatewayStatus = "FAULTY"
	GatewayStatusCircuitBreakerOpen GatewayStatus = "CIRCUIT_BREAKER_OPEN"
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

// JSONBMap JSONB 类型映射
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

// InstanceMetricsEntity API实例指标实体
type InstanceMetricsEntity struct {
	ID                   string        `gorm:"column:id;primaryKey" json:"id"`
	RegistryID           string        `gorm:"column:registry_id" json:"registry_id"`
	TimestampWindow      time.Time     `gorm:"column:timestamp_window" json:"timestamp_window"`
	SuccessCount         int64         `gorm:"column:success_count" json:"success_count"`
	FailureCount         int64         `gorm:"column:failure_count" json:"failure_count"`
	TotalLatencyMs       int64         `gorm:"column:total_latency_ms" json:"total_latency_ms"`
	Concurrency          int           `gorm:"column:concurrency" json:"concurrency"`
	CurrentGatewayStatus GatewayStatus `gorm:"column:current_gateway_status" json:"current_gateway_status"`
	LastReportedAt       time.Time     `gorm:"column:last_reported_at" json:"last_reported_at"`
	AdditionalMetrics    JSONBMap      `gorm:"column:additional_metrics;type:jsonb" json:"additional_metrics"`
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

// GetSuccessRate 计算成功率
func (m *InstanceMetricsEntity) GetSuccessRate() float64 {
	total := m.SuccessCount + m.FailureCount
	if total == 0 {
		return 1.0
	}
	return float64(m.SuccessCount) / float64(total)
}

// GetAverageLatency 计算平均延迟
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
