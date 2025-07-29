package validator

import (
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator 验证器
type Validator struct {
	errors ValidationErrors
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// AddError 添加验证错误
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// HasErrors 检查是否有错误
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors 获取所有错误
func (v *Validator) GetErrors() ValidationErrors {
	return v.errors
}

// Clear 清空错误
func (v *Validator) Clear() {
	v.errors = make(ValidationErrors, 0)
}

// Required 必填验证
func (v *Validator) Required(field, value string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "字段不能为空")
	}
	return v
}

// MinLength 最小长度验证
func (v *Validator) MinLength(field, value string, min int) *Validator {
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("长度不能少于%d个字符", min))
	}
	return v
}

// MaxLength 最大长度验证
func (v *Validator) MaxLength(field, value string, max int) *Validator {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("长度不能超过%d个字符", max))
	}
	return v
}

// Email 邮箱验证
func (v *Validator) Email(field, value string) *Validator {
	if value != "" {
		if _, err := mail.ParseAddress(value); err != nil {
			v.AddError(field, "邮箱格式不正确")
		}
	}
	return v
}

// Pattern 正则表达式验证
func (v *Validator) Pattern(field, value, pattern, message string) *Validator {
	if value != "" {
		matched, err := regexp.MatchString(pattern, value)
		if err != nil || !matched {
			v.AddError(field, message)
		}
	}
	return v
}

// Username 用户名验证
func (v *Validator) Username(field, value string) *Validator {
	if value == "" {
		return v
	}

	// 用户名只能包含字母、数字、下划线，长度3-20
	pattern := `^[a-zA-Z0-9_]{3,20}$`
	return v.Pattern(field, value, pattern, "用户名只能包含字母、数字、下划线，长度3-20个字符")
}

// Password 密码强度验证
func (v *Validator) Password(field, value string) *Validator {
	if value == "" {
		return v
	}

	// 密码长度至少8位
	if len(value) < 8 {
		v.AddError(field, "密码长度至少8位")
		return v
	}

	// 检查密码复杂度
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range value {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		v.AddError(field, "密码必须包含大写字母")
	}
	if !hasLower {
		v.AddError(field, "密码必须包含小写字母")
	}
	if !hasDigit {
		v.AddError(field, "密码必须包含数字")
	}
	if !hasSpecial {
		v.AddError(field, "密码必须包含特殊字符")
	}

	return v
}

// URL URL验证
func (v *Validator) URL(field, value string) *Validator {
	if value == "" {
		return v
	}

	// 简单的URL格式验证
	pattern := `^https?://[\w\-]+(\.[\w\-]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?$`
	return v.Pattern(field, value, pattern, "URL格式不正确")
}

// IP IP地址验证
func (v *Validator) IP(field, value string) *Validator {
	if value == "" {
		return v
	}

	// IPv4格式验证
	pattern := `^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	return v.Pattern(field, value, pattern, "IP地址格式不正确")
}

// Port 端口验证
func (v *Validator) Port(field, value string) *Validator {
	if value == "" {
		return v
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		v.AddError(field, "端口号必须在1-65535之间")
	}
	return v
}

// JSON JSON格式验证
func (v *Validator) JSON(field, value string) *Validator {
	if value == "" {
		return v
	}

	// 简单的JSON格式检查
	value = strings.TrimSpace(value)
	if !((strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}")) ||
		(strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"))) {
		v.AddError(field, "JSON格式不正确")
	}
	return v
}

// In 枚举值验证
func (v *Validator) In(field, value string, allowed []string) *Validator {
	if value == "" {
		return v
	}

	for _, item := range allowed {
		if value == item {
			return v
		}
	}

	v.AddError(field, fmt.Sprintf("值必须是以下之一: %s", strings.Join(allowed, ", ")))
	return v
}

// Range 数值范围验证
func (v *Validator) Range(field, value string, min, max int) *Validator {
	if value == "" {
		return v
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "必须是有效的数字")
		return v
	}

	if num < min || num > max {
		v.AddError(field, fmt.Sprintf("数值必须在%d-%d之间", min, max))
	}
	return v
}

// SQLInjectionCheck SQL注入检查
func (v *Validator) SQLInjectionCheck(field, value string) *Validator {
	if value == "" {
		return v
	}

	// 检查常见的SQL注入模式
	dangerousPatterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)\s`,
		`(?i)(script|javascript|vbscript|onload|onerror|onclick)`,
		`(?i)(\-\-|\/\*|\*\/|;|\||&)`,
		`(?i)(char|nchar|varchar|nvarchar|ascii|substring)\s*\(`,
		`(?i)(waitfor|delay|benchmark|sleep)\s*\(`,
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "输入包含潜在的安全风险字符")
			break
		}
	}

	return v
}

// XSSCheck XSS攻击检查
func (v *Validator) XSSCheck(field, value string) *Validator {
	if value == "" {
		return v
	}

	// 检查常见的XSS模式
	dangerousPatterns := []string{
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)<iframe[^>]*>.*?</iframe>`,
		`(?i)<object[^>]*>.*?</object>`,
		`(?i)<embed[^>]*>`,
		`(?i)<link[^>]*>`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)on\w+\s*=`,
	}

	for _, pattern := range dangerousPatterns {
		matched, _ := regexp.MatchString(pattern, value)
		if matched {
			v.AddError(field, "输入包含潜在的XSS攻击代码")
			break
		}
	}

	return v
}

// SanitizeInput 清理输入
func SanitizeInput(input string) string {
	// 移除危险字符
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#x27;")
	input = strings.ReplaceAll(input, "/", "&#x2F;")

	return strings.TrimSpace(input)
}

// ValidateStruct 结构体验证接口
type ValidateStruct interface {
	Validate() error
}

// ValidateRequest 验证请求结构体
func ValidateRequest(req ValidateStruct) error {
	return req.Validate()
}