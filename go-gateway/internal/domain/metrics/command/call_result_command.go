package command

// CallResultCommand 调用结果命令
type CallResultCommand struct {
	ProjectID    string                 `json:"project_id"`
	InstanceID   string                 `json:"instance_id"`
	Success      bool                   `json:"success"`
	LatencyMs    int64                  `json:"latency_ms"`
	ErrorMessage string                 `json:"error_message"`
	ErrorType    string                 `json:"error_type"`
	UsageMetrics map[string]interface{} `json:"usage_metrics"`
	CallTimestamp int64                 `json:"call_timestamp"`
}
