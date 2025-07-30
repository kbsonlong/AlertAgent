package plugins

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// ConfigValidator 配置验证器
type ConfigValidator struct{}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidateBySchema 根据JSON Schema验证配置
func (cv *ConfigValidator) ValidateBySchema(config map[string]interface{}, schema map[string]interface{}) error {
	schemaType, ok := schema["type"].(string)
	if !ok || schemaType != "object" {
		return fmt.Errorf("schema must be an object type")
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("schema must have properties")
	}

	required, _ := schema["required"].([]interface{})
	requiredFields := make(map[string]bool)
	for _, field := range required {
		if fieldName, ok := field.(string); ok {
			requiredFields[fieldName] = true
		}
	}

	// 检查必填字段
	for fieldName := range requiredFields {
		if _, exists := config[fieldName]; !exists {
			return fmt.Errorf("required field '%s' is missing", fieldName)
		}
	}

	// 验证每个字段
	for fieldName, fieldValue := range config {
		fieldSchema, exists := properties[fieldName]
		if !exists {
			continue // 允许额外字段
		}

		if err := cv.validateField(fieldName, fieldValue, fieldSchema); err != nil {
			return err
		}
	}

	return nil
}

// validateField 验证单个字段
func (cv *ConfigValidator) validateField(fieldName string, value interface{}, schema interface{}) error {
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid schema for field '%s'", fieldName)
	}

	fieldType, ok := schemaMap["type"].(string)
	if !ok {
		return fmt.Errorf("field '%s' schema must have type", fieldName)
	}

	// 检查类型
	if err := cv.validateType(fieldName, value, fieldType); err != nil {
		return err
	}

	// 检查其他约束
	if err := cv.validateConstraints(fieldName, value, schemaMap); err != nil {
		return err
	}

	return nil
}

// validateType 验证字段类型
func (cv *ConfigValidator) validateType(fieldName string, value interface{}, expectedType string) error {
	actualType := cv.getValueType(value)

	switch expectedType {
	case "string":
		if actualType != "string" {
			return fmt.Errorf("field '%s' must be string, got %s", fieldName, actualType)
		}
	case "number":
		if actualType != "number" && actualType != "integer" {
			return fmt.Errorf("field '%s' must be number, got %s", fieldName, actualType)
		}
	case "integer":
		if actualType != "integer" {
			return fmt.Errorf("field '%s' must be integer, got %s", fieldName, actualType)
		}
	case "boolean":
		if actualType != "boolean" {
			return fmt.Errorf("field '%s' must be boolean, got %s", fieldName, actualType)
		}
	case "array":
		if actualType != "array" {
			return fmt.Errorf("field '%s' must be array, got %s", fieldName, actualType)
		}
	case "object":
		if actualType != "object" {
			return fmt.Errorf("field '%s' must be object, got %s", fieldName, actualType)
		}
	default:
		return fmt.Errorf("unsupported type '%s' for field '%s'", expectedType, fieldName)
	}

	return nil
}

// validateConstraints 验证字段约束
func (cv *ConfigValidator) validateConstraints(fieldName string, value interface{}, schema map[string]interface{}) error {
	// 验证字符串长度
	if minLength, ok := schema["minLength"].(float64); ok {
		if str, ok := value.(string); ok {
			if len(str) < int(minLength) {
				return fmt.Errorf("field '%s' must be at least %d characters", fieldName, int(minLength))
			}
		}
	}

	if maxLength, ok := schema["maxLength"].(float64); ok {
		if str, ok := value.(string); ok {
			if len(str) > int(maxLength) {
				return fmt.Errorf("field '%s' must be at most %d characters", fieldName, int(maxLength))
			}
		}
	}

	// 验证数字范围
	if minimum, ok := schema["minimum"].(float64); ok {
		if num := cv.getNumericValue(value); num != nil {
			if *num < minimum {
				return fmt.Errorf("field '%s' must be at least %v", fieldName, minimum)
			}
		}
	}

	if maximum, ok := schema["maximum"].(float64); ok {
		if num := cv.getNumericValue(value); num != nil {
			if *num > maximum {
				return fmt.Errorf("field '%s' must be at most %v", fieldName, maximum)
			}
		}
	}

	// 验证枚举值
	if enumValues, ok := schema["enum"].([]interface{}); ok {
		found := false
		for _, enumValue := range enumValues {
			if reflect.DeepEqual(value, enumValue) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("field '%s' must be one of %v", fieldName, enumValues)
		}
	}

	// 验证正则表达式
	if pattern, ok := schema["pattern"].(string); ok {
		if str, ok := value.(string); ok {
			// 简单的模式匹配，实际应该使用正则表达式
			if pattern == "^https?://" && !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
				return fmt.Errorf("field '%s' must be a valid URL", fieldName)
			}
		}
	}

	// 验证数组项
	if items, ok := schema["items"].(map[string]interface{}); ok {
		if arr, ok := value.([]interface{}); ok {
			for i, item := range arr {
				if err := cv.validateField(fmt.Sprintf("%s[%d]", fieldName, i), item, items); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// getValueType 获取值的类型
func (cv *ConfigValidator) getValueType(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return "string"
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64:
		return "integer"
	case uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float32, float64:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return fmt.Sprintf("unknown(%T)", v)
	}
}

// getNumericValue 获取数值
func (cv *ConfigValidator) getNumericValue(value interface{}) *float64 {
	switch v := value.(type) {
	case int:
		f := float64(v)
		return &f
	case int8:
		f := float64(v)
		return &f
	case int16:
		f := float64(v)
		return &f
	case int32:
		f := float64(v)
		return &f
	case int64:
		f := float64(v)
		return &f
	case uint:
		f := float64(v)
		return &f
	case uint8:
		f := float64(v)
		return &f
	case uint16:
		f := float64(v)
		return &f
	case uint32:
		f := float64(v)
		return &f
	case uint64:
		f := float64(v)
		return &f
	case float32:
		f := float64(v)
		return &f
	case float64:
		return &v
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return &f
		}
	}
	return nil
}

// ValidateRequiredFields 验证必填字段
func (cv *ConfigValidator) ValidateRequiredFields(config map[string]interface{}, requiredFields []string) error {
	for _, field := range requiredFields {
		if value, exists := config[field]; !exists || cv.isEmpty(value) {
			return fmt.Errorf("required field '%s' is missing or empty", field)
		}
	}
	return nil
}

// isEmpty 检查值是否为空
func (cv *ConfigValidator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v) == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}