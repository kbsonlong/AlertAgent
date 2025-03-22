package types

import "time"

// AlertTask 告警任务
type AlertTask struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Source    string    `json:"source"`
	Content   string    `json:"content"`
	RuleID    uint      `json:"rule_id"`
	GroupID   uint      `json:"group_id"`
}

// AlertResult 告警处理结果
type AlertResult struct {
	TaskID    uint      `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
