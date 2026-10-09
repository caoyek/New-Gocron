package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/caoyek/New-Gocron/internal/modules/logger"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"github.com/jakecoffman/cron"
)

func TestHTTPTaskOptions(t *testing.T) {
	for _, tc := range []struct {
		name                                                        string
		method                                                      models.TaskHTTPMethod
		mode, body, query, expectedBody, expectedQuery, contentType string
	}{
		{"get", models.TaskHTTPMethodGet, "", "", "a=1", "", "a=1", ""},
		{"legacy post", models.TaskHttpMethodPost, "", "", "a=1&b=two?three", "a=1&b=two?three", "", "application/x-www-form-urlencoded"},
		{"json post", models.TaskHttpMethodPost, "json", `{"report":true}`, "token=1", `{"report":true}`, "token=1", "application/json"},
		{"no body", models.TaskHttpMethodPost, "json", "", "token=1", "", "token=1", ""},
		{"whitespace body", models.TaskHttpMethodPost, "json", " \n\t ", "token=1", "", "token=1", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := ioutil.ReadAll(r.Body)
				if string(body) != tc.expectedBody || r.URL.RawQuery != tc.expectedQuery || r.Header.Get("Content-Type") != tc.contentType || r.Header.Get("Authorization") != "Bearer test" || r.UserAgent() != "custom-client" {
					t.Errorf("request mismatch: body=%s query=%s headers=%v", body, r.URL.RawQuery, r.Header)
				}
				w.Write([]byte("done"))
			}))
			defer server.Close()
			result, err := new(HTTPHandler).Run(models.Task{Command: server.URL + "?" + tc.query, HttpMethod: tc.method, HttpBodyMode: tc.mode, HttpBody: tc.body, HttpHeaders: `{"Authorization":"Bearer test","User-Agent":"custom-client"}`, Timeout: 2}, 1)
			if err != nil || result != "done" {
				t.Fatalf("%s %v", result, err)
			}
		})
	}
}

func TestDisabledLogsStillExecute(t *testing.T) {
	// No database is configured: both hooks must avoid all database access.
	task := models.Task{Id: 1, LogDisabled: true, Multi: 1}
	id1, id2 := beforeExecJob(task), beforeExecJob(task)
	if id1 >= 0 || id2 >= 0 || id1 == id2 {
		t.Fatal("missing unique execution IDs")
	}
	afterExecJob(task, TaskResult{Result: "success"}, id1)
	runInstance.add(task.Id)
	defer runInstance.done(task.Id)
	task.Multi = 0
	if beforeExecJob(task) != 0 {
		t.Fatal("single-instance protection lost")
	}
}

// Only uniquely prefixed tables are touched, never the application's tables.
func TestMySQLTaskOptionsLifecycle(t *testing.T) {
	dsn := os.Getenv("GOCRON_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set GOCRON_MIGRATION_TEST_DSN for isolated MySQL integration")
	}
	logger.InitLogger()
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	prefix := fmt.Sprintf("task_options_test_%d_", time.Now().UnixNano())
	engine.SetTableMapper(core.NewPrefixMapper(core.SnakeMapper{}, prefix))
	oldDB, oldPrefix, oldCron := models.Db, models.TablePrefix, serviceCron
	models.Db, models.TablePrefix, serviceCron = engine, prefix, cron.New()
	defer func() { models.Db, models.TablePrefix, serviceCron = oldDB, oldPrefix, oldCron }()
	if err := engine.Sync2(new(models.Task), new(models.TaskLog)); err != nil {
		t.Fatal(err)
	}
	defer engine.DropTables(new(models.Task), new(models.TaskLog))
	legacy := models.Task{Name: "legacy", Status: models.Enabled, Protocol: models.TaskHTTP, Spec: "0 0 0 * * *"}
	if _, err := engine.Insert(&legacy); err != nil {
		t.Fatal(err)
	}
	for _, column := range []string{"http_headers", "http_body_mode", "http_body", "expires_at", "log_disabled", "log_retention_days"} {
		if _, err := engine.Exec("ALTER TABLE `" + prefix + "task` DROP COLUMN `" + column + "`"); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := new(models.Migration).EnsureTaskOptionsSchema(); err != nil {
			t.Fatal(err)
		}
	}
	var migrated models.Task
	if _, err := engine.ID(legacy.Id).Get(&migrated); err != nil {
		t.Fatal(err)
	}
	if migrated.Name != "legacy" || migrated.LogDisabled || migrated.LogRetentionDays != 0 || migrated.ExpiresAt != "" {
		t.Fatalf("legacy defaults lost: %+v", migrated)
	}

	// Save/read all options, including zero and false after previously nonzero values.
	configured := migrated
	configured.HttpHeaders, configured.HttpBodyMode, configured.HttpBody = `{"X-Test":"yes"}`, "json", `{"ok":true}`
	configured.LogDisabled, configured.LogRetentionDays = true, 7
	configured.ExpiresAt = time.Now().Add(-time.Hour).Format(models.DefaultTimeFormat)
	if _, err := configured.UpdateBean(legacy.Id); err != nil {
		t.Fatal(err)
	}
	var loaded models.Task
	if _, err := engine.ID(legacy.Id).Get(&loaded); err != nil {
		t.Fatal(err)
	}
	if !loaded.LogDisabled || loaded.LogRetentionDays != 7 || loaded.HttpBody != configured.HttpBody {
		t.Fatal("options roundtrip failed")
	}
	ServiceTask.pauseExpiredTasks()
	var paused models.Task
	engine.ID(legacy.Id).Get(&paused)
	if paused.Status != models.Disabled || beforeExecJob(loaded) != 0 {
		t.Fatal("expired task was not paused/blocked")
	}
	configured.ExpiresAt, configured.LogDisabled, configured.LogRetentionDays = "", false, 0
	if _, err := configured.UpdateBean(legacy.Id); err != nil {
		t.Fatal(err)
	}
	var restored models.Task
	engine.ID(legacy.Id).Get(&restored)
	if restored.ExpiresAt != "" || restored.LogDisabled || restored.LogRetentionDays != 0 {
		t.Fatal("clearing options failed")
	}
	// A stale expiration snapshot must not pause a task whose date was extended.
	engine.Table(new(models.Task)).ID(legacy.Id).Update(models.CommonMap{"status": models.Enabled})
	ServiceTask.PauseExpired(loaded)
	engine.ID(legacy.Id).Get(&restored)
	if restored.Status != models.Enabled {
		t.Fatal("stale expiration paused a renewed task")
	}

	retained := models.Task{Name: "retention", LogRetentionDays: 7}
	engine.Insert(&retained)
	for _, item := range []struct {
		taskID int
		age    int
		status models.Status
	}{
		{retained.Id, 10, models.Finish}, {retained.Id, 1, models.Finish},
		{retained.Id, 10, models.Running}, {legacy.Id, 10, models.Finish},
	} {
		log := models.TaskLog{TaskId: item.taskID, Status: item.status}
		if _, err := log.Create(); err != nil {
			t.Fatal(err)
		}
		// Use raw SQL to set fixtures without automatic timestamp updates.
		if _, err := engine.Exec("UPDATE `"+prefix+"task_log` SET end_time = ? WHERE id = ?", time.Now().AddDate(0, 0, -item.age), log.Id); err != nil {
			t.Fatal(err)
		}
	}
	var candidates []models.TaskLog
	if err := engine.Where("task_id = ?", retained.Id).Find(&candidates); err != nil {
		t.Fatal(err)
	}
	ids := make([]int64, len(candidates))
	for i, entry := range candidates {
		ids[i] = entry.Id
	}
	for _, days := range []int{0, 90} {
		if _, err := engine.Table(new(models.Task)).ID(retained.Id).Update(models.CommonMap{"log_retention_days": days}); err != nil {
			t.Fatal(err)
		}
		if affected, err := deleteExpiredLogBatch(retained, time.Now().AddDate(0, 0, -7), ids); err != nil || affected != 0 {
			t.Fatalf("stale retention deleted logs after change to %d: %d %v", days, affected, err)
		}
	}
	if count, err := engine.Count(new(models.TaskLog)); err != nil || count != 4 {
		t.Fatalf("changed retention lost logs: %d %v", count, err)
	}
	if _, err := engine.Table(new(models.Task)).ID(retained.Id).Update(models.CommonMap{"log_retention_days": 7}); err != nil {
		t.Fatal(err)
	}
	ServiceTask.cleanupLogs(make(chan struct{}))
	count, err := engine.Count(new(models.TaskLog))
	if err != nil || count != 3 {
		t.Fatalf("retention removed wrong records: %d %v", count, err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) }))
	defer server.Close()
	concurrencyQueue = ConcurrencyQueue{queue: make(chan struct{}, 1)}
	job := models.Task{Id: legacy.Id, Protocol: models.TaskHTTP, HttpMethod: models.TaskHTTPMethodGet, Command: server.URL, Multi: 1, LogDisabled: true}
	createJob(job)()
	afterDisabled, _ := engine.Count(new(models.TaskLog))
	if afterDisabled != count {
		t.Fatal("disabled task wrote a log")
	}
	job.LogDisabled = false
	createJob(job)()
	var result models.TaskLog
	engine.Where("task_id = ?", legacy.Id).Desc("id").Get(&result)
	if result.Result != "ok" || result.Status != models.Finish {
		t.Fatalf("enabled logging failed: %+v", result)
	}
}
