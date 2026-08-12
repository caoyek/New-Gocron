package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const notificationRuleVersion = 1

type NotificationRuleSet struct {
	Format  string             `json:"format"`
	Version int                `json:"version"`
	Mode    string             `json:"mode"`
	Rules   []NotificationRule `json:"rules"`
}

type NotificationRule struct {
	Type          string  `json:"type"`
	Value         string  `json:"value,omitempty"`
	CaseSensitive bool    `json:"case_sensitive,omitempty"`
	Field         string  `json:"field,omitempty"`
	Operator      string  `json:"operator,omitempty"`
	Number        float64 `json:"number,omitempty"`
}

func ParseNotificationRules(value string) (NotificationRuleSet, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return NotificationRuleSet{}, errors.New("通知匹配规则不能为空")
	}
	if len(value) > 60000 {
		return NotificationRuleSet{}, errors.New("通知匹配规则内容过长")
	}

	legacyRules := func() NotificationRuleSet {
		return NotificationRuleSet{
			Format:  "notification_rules",
			Version: notificationRuleVersion,
			Mode:    "any",
			Rules: []NotificationRule{{
				Type:          "contains",
				Value:         value,
				CaseSensitive: true,
			}},
		}
	}
	if !strings.HasPrefix(value, "{") {
		return legacyRules(), nil
	}

	ruleSet := NotificationRuleSet{}
	if err := json.Unmarshal([]byte(value), &ruleSet); err != nil {
		return legacyRules(), nil
	}
	if ruleSet.Format != "notification_rules" {
		return legacyRules(), nil
	}
	if err := ValidateNotificationRules(ruleSet); err != nil {
		return NotificationRuleSet{}, err
	}

	return ruleSet, nil
}

func ValidateNotificationRules(ruleSet NotificationRuleSet) error {
	if ruleSet.Format != "notification_rules" {
		return errors.New("通知匹配规则格式错误")
	}
	if ruleSet.Version != notificationRuleVersion {
		return errors.New("不支持的通知匹配规则版本")
	}
	if ruleSet.Mode != "any" && ruleSet.Mode != "all" {
		return errors.New("通知匹配方式必须是满足任意规则或满足全部规则")
	}
	if len(ruleSet.Rules) == 0 {
		return errors.New("至少添加一条通知匹配规则")
	}
	if len(ruleSet.Rules) > 20 {
		return errors.New("通知匹配规则不能超过20条")
	}

	for i, rule := range ruleSet.Rules {
		if err := validateNotificationRule(rule); err != nil {
			return fmt.Errorf("第%d条规则: %s", i+1, err.Error())
		}
	}

	return nil
}

func validateNotificationRule(rule NotificationRule) error {
	switch rule.Type {
	case "contains", "not_contains", "wildcard":
		if strings.TrimSpace(rule.Value) == "" {
			return errors.New("匹配内容不能为空")
		}
		if len(rule.Value) > 1000 {
			return errors.New("匹配内容不能超过1000个字符")
		}
	case "regex":
		if strings.TrimSpace(rule.Value) == "" {
			return errors.New("正则表达式不能为空")
		}
		if _, err := compileRuleRegexp(rule.Value, rule.CaseSensitive); err != nil {
			return fmt.Errorf("正则表达式无效: %s", err.Error())
		}
		if len(rule.Value) > 1000 {
			return errors.New("正则表达式不能超过1000个字符")
		}
	case "number":
		if strings.TrimSpace(rule.Field) == "" {
			return errors.New("数值字段不能为空")
		}
		if !validNumberOperator(rule.Operator) {
			return errors.New("数值比较运算符无效")
		}
		if len(rule.Field) > 128 {
			return errors.New("数值字段不能超过128个字符")
		}
	default:
		return errors.New("规则类型无效")
	}

	return nil
}

func MatchNotificationRules(output, value string) (bool, error) {
	ruleSet, err := ParseNotificationRules(value)
	if err != nil {
		return false, err
	}

	if ruleSet.Mode == "all" {
		for _, rule := range ruleSet.Rules {
			matched, err := matchNotificationRule(output, rule)
			if err != nil || !matched {
				return false, err
			}
		}
		return true, nil
	}

	for _, rule := range ruleSet.Rules {
		matched, err := matchNotificationRule(output, rule)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

func matchNotificationRule(output string, rule NotificationRule) (bool, error) {
	switch rule.Type {
	case "contains", "not_contains":
		haystack, needle := output, rule.Value
		if !rule.CaseSensitive {
			haystack = strings.ToLower(haystack)
			needle = strings.ToLower(needle)
		}
		matched := strings.Contains(haystack, needle)
		if rule.Type == "not_contains" {
			matched = !matched
		}
		return matched, nil
	case "wildcard":
		pattern := regexp.QuoteMeta(rule.Value)
		pattern = strings.Replace(pattern, `\*`, `.*`, -1)
		pattern = strings.Replace(pattern, `\?`, `.`, -1)
		pattern = "(?s)" + pattern
		re, err := compileRuleRegexp(pattern, rule.CaseSensitive)
		if err != nil {
			return false, err
		}
		return re.MatchString(output), nil
	case "regex":
		re, err := compileRuleRegexp(rule.Value, rule.CaseSensitive)
		if err != nil {
			return false, err
		}
		return re.MatchString(output), nil
	case "number":
		return matchNumberRule(output, rule)
	}

	return false, errors.New("规则类型无效")
}

func compileRuleRegexp(pattern string, caseSensitive bool) (*regexp.Regexp, error) {
	if !caseSensitive {
		pattern = "(?i)" + pattern
	}
	return regexp.Compile(pattern)
}

func matchNumberRule(output string, rule NotificationRule) (bool, error) {
	field := regexp.QuoteMeta(strings.TrimSpace(rule.Field))
	pattern := field + `\s*(?:[:：=]\s*)?(-?(?:\d+(?:\.\d+)?|\.\d+))`
	re, err := compileRuleRegexp(pattern, rule.CaseSensitive)
	if err != nil {
		return false, err
	}

	for _, match := range re.FindAllStringSubmatch(output, -1) {
		value, err := strconv.ParseFloat(match[1], 64)
		if err == nil && compareNotificationNumber(value, rule.Operator, rule.Number) {
			return true, nil
		}
	}

	return false, nil
}

func validNumberOperator(operator string) bool {
	switch operator {
	case ">", ">=", "<", "<=", "=", "!=":
		return true
	}
	return false
}

func compareNotificationNumber(actual float64, operator string, expected float64) bool {
	switch operator {
	case ">":
		return actual > expected
	case ">=":
		return actual >= expected
	case "<":
		return actual < expected
	case "<=":
		return actual <= expected
	case "=":
		return actual == expected
	case "!=":
		return actual != expected
	}
	return false
}
