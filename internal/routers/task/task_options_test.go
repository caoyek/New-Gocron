package task

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/go-macaron/binding"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
)

func TestMySQLTaskOptionsAPI(t *testing.T) {
	dsn := os.Getenv("GOCRON_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set GOCRON_MIGRATION_TEST_DSN for isolated MySQL integration")
	}
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	prefix := fmt.Sprintf("task_options_api_test_%d_", time.Now().UnixNano())
	engine.SetTableMapper(core.NewPrefixMapper(core.SnakeMapper{}, prefix))
	oldDB, oldPrefix := models.Db, models.TablePrefix
	models.Db, models.TablePrefix = engine, prefix
	defer func() { models.Db, models.TablePrefix = oldDB, oldPrefix }()
	tables := []interface{}{new(models.Task), new(models.TaskLog), new(models.TaskHost), new(models.Host)}
	if err := engine.Sync2(tables...); err != nil {
		t.Fatal(err)
	}
	defer engine.DropTables(tables...)
	item := models.Task{Name: "api-options", Status: models.Disabled}
	if _, err := engine.Insert(&item); err != nil {
		t.Fatal(err)
	}
	m := macaron.New()
	m.Post("/task/store", binding.Bind(TaskForm{}), Store)
	m.Get("/task/:id", Detail)
	m.Get("/task/run/:id", Run)
	m.Post("/task/enable/:id", Enable)
	values := url.Values{
		"id": {strconv.Itoa(item.Id)}, "name": {item.Name}, "level": {"1"}, "spec": {"0 0 0 * * *"},
		"dependency_status": {"1"}, "protocol": {"1"}, "http_method": {"2"}, "command": {"https://example.invalid/report"},
		"multi": {"2"}, "notify_status": {"1"}, "notify_type": {"2"},
		"http_headers": {`{"X-Test":"yes"}`}, "http_body_mode": {"json"}, "http_body": {`{"report":true}`},
		"expires_at": {time.Now().Add(time.Hour).Format(models.DefaultTimeFormat)}, "log_disabled": {"true"}, "log_retention_days": {"90"},
	}
	request := func(method, path string, body url.Values) struct {
		Code    int
		Data    models.Task
		Message string
	} { req := httptest.NewRequest(method, path, strings.NewReader(body.Encode())); req.Header.Set("Content-Type", "application/x-www-form-urlencoded"); response := httptest.NewRecorder(); m.ServeHTTP(response, req); var result struct {
		Code    int
		Data    models.Task
		Message string
	}; if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("invalid response: %s", response.Body.String())
	}; return result }
	if result := request("POST", "/task/store", values); result.Code != 0 {
		t.Fatal(result.Message)
	}
	result := request("GET", fmt.Sprintf("/task/%d", item.Id), nil)
	if !result.Data.LogDisabled || result.Data.LogRetentionDays != 90 || result.Data.HttpBody != values.Get("http_body") || result.Data.HttpHeaders != values.Get("http_headers") || result.Data.ExpiresAt != values.Get("expires_at") {
		t.Fatalf("save/read options mismatch: %+v", result.Data)
	}
	values.Set("http_body", "invalid json")
	if request("POST", "/task/store", values).Code == 0 {
		t.Fatal("invalid JSON accepted")
	}
	values.Set("http_body", `{"report":true}`)
	values.Set("expires_at", time.Now().Add(-time.Hour).Format(models.DefaultTimeFormat))
	if result := request("POST", "/task/store", values); result.Code != 0 {
		t.Fatal(result.Message)
	}
	for _, action := range []struct{ method, path string }{
		{"POST", fmt.Sprintf("/task/enable/%d", item.Id)}, {"GET", fmt.Sprintf("/task/run/%d", item.Id)},
	} {
		if result := request(action.method, action.path, nil); result.Code == 0 {
			t.Fatal("expired task action allowed")
		}
	}
	values.Set("expires_at", "")
	values.Set("log_disabled", "false")
	values.Set("log_retention_days", "0")
	if result := request("POST", "/task/store", values); result.Code != 0 {
		t.Fatal(result.Message)
	}
	cleared := request("GET", fmt.Sprintf("/task/%d", item.Id), nil).Data
	if cleared.LogDisabled || cleared.LogRetentionDays != 0 || cleared.ExpiresAt != "" {
		t.Fatal("zero/false clearing failed")
	}
}
