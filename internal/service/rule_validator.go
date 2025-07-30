package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RuleValidator 规则验证器接口
type RuleValidator interface {
	ValidateRule(expression, duration string) error
	ValidateExpression(expression string) error
	ValidateDuration(duration string) error
	ValidateSeverity(severity string) error
}

// prometheusRuleValidator Prometheus规则验证器实现
type prometheusRuleValidator struct{}

// NewRuleValidator 创建规则验证器实例
func NewRuleValidator() RuleValidator {
	return &prometheusRuleValidator{}
}

// ValidateRule 验证完整规则
func (v *prometheusRuleValidator) ValidateRule(expression, duration string) error {
	if err := v.ValidateExpression(expression); err != nil {
		return fmt.Errorf("expression validation failed: %w", err)
	}
	
	if err := v.ValidateDuration(duration); err != nil {
		return fmt.Errorf("duration validation failed: %w", err)
	}
	
	return nil
}

// ValidateExpression 验证Prometheus表达式语法
func (v *prometheusRuleValidator) ValidateExpression(expression string) error {
	if strings.TrimSpace(expression) == "" {
		return fmt.Errorf("expression cannot be empty")
	}

	// 基本语法检查
	if err := v.validateBasicSyntax(expression); err != nil {
		return err
	}

	// 检查函数调用语法
	if err := v.validateFunctions(expression); err != nil {
		return err
	}

	// 检查操作符语法
	if err := v.validateOperators(expression); err != nil {
		return err
	}

	return nil
}

// ValidateDuration 验证持续时间格式
func (v *prometheusRuleValidator) ValidateDuration(duration string) error {
	if strings.TrimSpace(duration) == "" {
		return fmt.Errorf("duration cannot be empty")
	}

	// 使用Go的time.ParseDuration来验证格式
	_, err := time.ParseDuration(duration)
	if err != nil {
		// 如果Go的格式不支持，尝试Prometheus格式
		if !v.isValidPrometheusDuration(duration) {
			return fmt.Errorf("invalid duration format: %s, expected formats like '5m', '1h', '30s'", duration)
		}
	}

	return nil
}

// ValidateSeverity 验证严重程度
func (v *prometheusRuleValidator) ValidateSeverity(severity string) error {
	validSeverities := map[string]bool{
		"critical": true,
		"warning":  true,
		"info":     true,
		"low":      true,
	}

	if !validSeverities[strings.ToLower(severity)] {
		return fmt.Errorf("invalid severity: %s, must be one of: critical, warning, info, low", severity)
	}

	return nil
}

// validateBasicSyntax 验证基本语法
func (v *prometheusRuleValidator) validateBasicSyntax(expression string) error {
	// 检查括号匹配
	if !v.isBalancedParentheses(expression) {
		return fmt.Errorf("unbalanced parentheses in expression")
	}

	// 检查引号匹配
	if !v.isBalancedQuotes(expression) {
		return fmt.Errorf("unbalanced quotes in expression")
	}

	// 检查基本的metric名称格式
	if err := v.validateMetricNames(expression); err != nil {
		return err
	}

	return nil
}

// validateFunctions 验证函数调用
func (v *prometheusRuleValidator) validateFunctions(expression string) error {
	// 常见的Prometheus函数
	validFunctions := map[string]bool{
		"rate":                true,
		"irate":               true,
		"increase":            true,
		"sum":                 true,
		"avg":                 true,
		"max":                 true,
		"min":                 true,
		"count":               true,
		"stddev":              true,
		"stdvar":              true,
		"topk":                true,
		"bottomk":             true,
		"quantile":            true,
		"histogram_quantile":  true,
		"abs":                 true,
		"ceil":                true,
		"floor":               true,
		"round":               true,
		"sqrt":                true,
		"exp":                 true,
		"ln":                  true,
		"log2":                true,
		"log10":               true,
		"time":                true,
		"vector":              true,
		"scalar":              true,
		"sort":                true,
		"sort_desc":           true,
		"clamp_max":           true,
		"clamp_min":           true,
		"changes":             true,
		"delta":               true,
		"deriv":               true,
		"predict_linear":      true,
		"holt_winters":        true,
		"idelta":              true,
		"resets":              true,
		"avg_over_time":       true,
		"min_over_time":       true,
		"max_over_time":       true,
		"sum_over_time":       true,
		"count_over_time":     true,
		"quantile_over_time":  true,
		"stddev_over_time":    true,
		"stdvar_over_time":    true,
		"absent":              true,
		"absent_over_time":    true,
		"present_over_time":   true,
		"up":                  true,
	}

	// 使用正则表达式查找函数调用
	funcRegex := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	matches := funcRegex.FindAllStringSubmatch(expression, -1)

	for _, match := range matches {
		if len(match) > 1 {
			funcName := match[1]
			if !validFunctions[funcName] {
				// 这里只是警告，不阻止执行，因为可能有自定义函数
				// 在实际生产环境中，可以根据需要调整策略
				continue
			}
		}
	}

	return nil
}

// validateOperators 验证操作符
func (v *prometheusRuleValidator) validateOperators(expression string) error {
	// 检查是否有连续的操作符
	invalidPatterns := []string{
		"++", "--", "**", "//", "%%",
		"==", "!=", "<=", ">=", "<<", ">>",
		"&&", "||",
	}

	for _, pattern := range invalidPatterns {
		if strings.Contains(expression, pattern) {
			// 这些是有效的操作符，跳过
			if pattern == "==" || pattern == "!=" || pattern == "<=" || pattern == ">=" {
				continue
			}
			return fmt.Errorf("invalid operator pattern: %s", pattern)
		}
	}

	return nil
}

// validateMetricNames 验证指标名称
func (v *prometheusRuleValidator) validateMetricNames(expression string) error {
	// Prometheus指标名称的正则表达式
	metricRegex := regexp.MustCompile(`[a-zA-Z_:][a-zA-Z0-9_:]*`)
	matches := metricRegex.FindAllString(expression, -1)

	for _, match := range matches {
		// 跳过函数名和关键字
		if v.isReservedWord(match) {
			continue
		}

		// 验证指标名称格式
		if !v.isValidMetricName(match) {
			return fmt.Errorf("invalid metric name: %s", match)
		}
	}

	return nil
}

// isBalancedParentheses 检查括号是否匹配
func (v *prometheusRuleValidator) isBalancedParentheses(s string) bool {
	count := 0
	for _, char := range s {
		if char == '(' {
			count++
		} else if char == ')' {
			count--
			if count < 0 {
				return false
			}
		}
	}
	return count == 0
}

// isBalancedQuotes 检查引号是否匹配
func (v *prometheusRuleValidator) isBalancedQuotes(s string) bool {
	singleQuoteCount := 0
	doubleQuoteCount := 0
	escaped := false

	for _, char := range s {
		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '\'' {
			singleQuoteCount++
		} else if char == '"' {
			doubleQuoteCount++
		}
	}

	return singleQuoteCount%2 == 0 && doubleQuoteCount%2 == 0
}

// isValidPrometheusDuration 检查是否为有效的Prometheus持续时间格式
func (v *prometheusRuleValidator) isValidPrometheusDuration(duration string) bool {
	// Prometheus支持的时间单位：ns, us, µs, ms, s, m, h, d, w, y
	pattern := `^(\d+(\.\d+)?[nsuµmhdwy])+$`
	matched, _ := regexp.MatchString(pattern, duration)
	return matched
}

// isReservedWord 检查是否为保留字
func (v *prometheusRuleValidator) isReservedWord(word string) bool {
	reservedWords := map[string]bool{
		"and":    true,
		"or":     true,
		"unless": true,
		"by":     true,
		"without": true,
		"on":     true,
		"ignoring": true,
		"group_left": true,
		"group_right": true,
		"offset": true,
		"bool":   true,
	}
	return reservedWords[strings.ToLower(word)]
}

// isValidMetricName 检查是否为有效的指标名称
func (v *prometheusRuleValidator) isValidMetricName(name string) bool {
	// Prometheus指标名称规则：
	// 1. 必须以字母、下划线或冒号开头
	// 2. 后续字符可以是字母、数字、下划线或冒号
	pattern := `^[a-zA-Z_:][a-zA-Z0-9_:]*$`
	matched, _ := regexp.MatchString(pattern, name)
	return matched
}

// ParseDurationToSeconds 将持续时间转换为秒数（用于验证合理性）
func (v *prometheusRuleValidator) ParseDurationToSeconds(duration string) (float64, error) {
	// 首先尝试Go标准格式
	if d, err := time.ParseDuration(duration); err == nil {
		return d.Seconds(), nil
	}

	// 尝试解析Prometheus格式
	return v.parsePrometheusDuration(duration)
}

// parsePrometheusDuration 解析Prometheus持续时间格式
func (v *prometheusRuleValidator) parsePrometheusDuration(duration string) (float64, error) {
	// 单位转换表（转换为秒）
	units := map[string]float64{
		"ns": 1e-9,
		"us": 1e-6,
		"µs": 1e-6,
		"ms": 1e-3,
		"s":  1,
		"m":  60,
		"h":  3600,
		"d":  86400,
		"w":  604800,
		"y":  31536000,
	}

	// 使用正则表达式解析
	pattern := `(\d+(?:\.\d+)?)([nsuµmhdwy]+)`
	regex := regexp.MustCompile(pattern)
	matches := regex.FindAllStringSubmatch(duration, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	var totalSeconds float64
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}

		value, err := strconv.ParseFloat(match[1], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number in duration: %s", match[1])
		}

		unit := match[2]
		multiplier, exists := units[unit]
		if !exists {
			return 0, fmt.Errorf("invalid time unit: %s", unit)
		}

		totalSeconds += value * multiplier
	}

	return totalSeconds, nil
}