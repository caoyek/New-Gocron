package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-xorm/core"
)

// Add only missing options; never rewrite unrelated legacy columns.
func (m *Migration) EnsureTaskOptionsSchema() error {
	session := Db.NewSession()
	defer session.Close()
	table := TablePrefix + "task"
	for _, column := range []struct{ name, definition string }{
		{"http_headers", "TEXT"},
		{"http_body_mode", "VARCHAR(16) NOT NULL DEFAULT ''"},
		{"http_body", "TEXT"},
		{"expires_at", "VARCHAR(19) NOT NULL DEFAULT ''"},
		{"log_disabled", "BOOLEAN NOT NULL DEFAULT FALSE"},
		{"log_retention_days", "INT NOT NULL DEFAULT 0"},
	} {
		var columnType string
		var err error
		switch Db.Dialect().DBType() {
		case core.MYSQL:
			columnType, _, err = mysqlColumnDefinition(session, table, column.name)
		case core.POSTGRES:
			columnType, _, err = postgresColumnDefinition(session, table, column.name)
		default:
			return errors.New("任务配置升级仅支持 MySQL 和 PostgreSQL")
		}
		if err != nil {
			return err
		}
		if columnType == "" {
			if _, err := session.Exec("ALTER TABLE " + Db.Quote(table) + " ADD COLUMN " + Db.Quote(column.name) + " " + column.definition); err != nil {
				return err
			}
		}
	}
	return nil
}

func (task Task) Expired(now time.Time) bool {
	if task.ExpiresAt == "" {
		return false
	}
	expires, err := time.ParseInLocation(DefaultTimeFormat, task.ExpiresAt, time.Local)
	return err != nil || !now.Before(expires)
}

func ParseHTTPHeaders(raw string) (map[string]string, error) {
	headers := make(map[string]string)
	if strings.TrimSpace(raw) == "" {
		return headers, nil
	}
	var values map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &values); err != nil || values == nil {
		return nil, errors.New("HTTP Header 必须为 JSON 对象，名称和值均为字符串")
	}
	for name, value := range values {
		var text string
		if len(value) == 0 || value[0] != '"' || json.Unmarshal(value, &text) != nil {
			return nil, errors.New("HTTP Header 的值必须是字符串")
		}
		headers[name] = text
	}
	seen := make(map[string]bool)
	for name, value := range headers {
		key := strings.ToLower(name)
		if seen[key] {
			return nil, errors.New("HTTP Header 名称不能重复（不区分大小写）")
		}
		seen[key] = true
		if name == "" {
			return nil, errors.New("HTTP Header 名称不能为空")
		}
		for _, ch := range name {
			if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", ch)) {
				return nil, errors.New("HTTP Header 名称包含无效字符")
			}
		}
		for _, ch := range value {
			if (ch < 32 && ch != '\t') || ch == 127 {
				return nil, errors.New("HTTP Header 值包含无效控制字符")
			}
		}
		switch strings.ToLower(name) {
		case "host", "content-length", "transfer-encoding":
			return nil, fmt.Errorf("HTTP Header 不支持手动设置 %s", name)
		}
	}
	return headers, nil
}

func (task Task) ValidateOptions() error {
	if task.ExpiresAt != "" {
		if _, err := time.ParseInLocation(DefaultTimeFormat, task.ExpiresAt, time.Local); err != nil {
			return errors.New("到期时间格式应为 YYYY-MM-DD HH:mm:ss")
		}
	}
	switch task.LogRetentionDays {
	case 0, 7, 30, 90, 180, 365:
	default:
		return errors.New("日志保留时间请选择 7天、30天、90天、180天、1年或永久")
	}
	if task.Protocol == TaskHTTP {
		if _, err := ParseHTTPHeaders(task.HttpHeaders); err != nil {
			return err
		}
		if task.HttpBodyMode != "" && task.HttpBodyMode != "json" {
			return errors.New("HTTP Body 类型无效")
		}
		if task.HttpBodyMode == "json" && (task.HttpMethod != TaskHttpMethodPost || (strings.TrimSpace(task.HttpBody) != "" && !json.Valid([]byte(task.HttpBody)))) {
			return errors.New("JSON Body 仅支持 POST 请求，内容必须是有效 JSON")
		}
	}
	return nil
}
