package request

// SelectInstanceRequest 选择API实例请求
type SelectInstanceRequest struct {
	UserID        string   `json:"userId"`
	ApiIdentifier string   `json:"apiIdentifier" binding:"required"`
	ApiType       string   `json:"apiType" binding:"required"`
	AffinityKey   string   `json:"affinityKey"`
	AffinityType  string   `json:"affinityType"`
	FallbackChain []string `json:"fallbackChain"`
}

// HasAffinityRequirement 检查是否有亲和性要求
func (r *SelectInstanceRequest) HasAffinityRequirement() bool {
	return r.AffinityKey != "" && r.AffinityType != ""
}

// HasFallbackChain 检查是否有降级链
func (r *SelectInstanceRequest) HasFallbackChain() bool {
	return len(r.FallbackChain) > 0
}

// ReportResultRequest 上报调用结果请求
type ReportResultRequest struct {
	UserID        string                 `json:"userId"`
	InstanceID    string                 `json:"instanceId" binding:"required"`
	BusinessID    string                 `json:"businessId" binding:"required"`
	Success       bool                   `json:"success" binding:"required"`
	LatencyMs     int64                  `json:"latencyMs" binding:"required"`
	ErrorMessage  string                 `json:"errorMessage"`
	ErrorType     string                 `json:"errorType"`
	UsageMetrics  map[string]interface{} `json:"usageMetrics"`
	CallTimestamp int64                  `json:"callTimestamp" binding:"required"`
}
