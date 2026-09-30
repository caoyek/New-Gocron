package models

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-xorm/xorm"
)

// Opt-in integration test. Only a uniquely named test table is modified.
func TestMySQLNotifyReceiverMigration(t *testing.T) {
	dsn := os.Getenv("GOCRON_MIGRATION_TEST_DSN")
	if dsn == "" {
		t.Skip("set GOCRON_MIGRATION_TEST_DSN to test against MySQL")
	}
	engine, err := xorm.NewEngine("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	session := engine.NewSession()
	defer session.Close()
	// Pin the session connection so strict mode applies to every statement.
	if err := session.Begin(); err != nil {
		t.Fatal(err)
	}
	defer session.Rollback()
	if _, err := session.Exec("SET SESSION sql_mode = 'STRICT_ALL_TABLES'"); err != nil {
		t.Fatal(err)
	}
	table := fmt.Sprintf("migration_receiver_test_%d", time.Now().UnixNano())
	if _, err := session.Exec("CREATE TABLE `" + table + "` (id INT PRIMARY KEY, notify_receiver_id INT NULL)"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := session.Exec("DROP TABLE `" + table + "`"); err != nil {
			t.Error(err)
		}
	}()
	if _, err := session.Exec("INSERT INTO `" + table + "` VALUES (1, NULL), (2, 8), (3, 0)"); err != nil {
		t.Fatal(err)
	}
	if err := ensureMySQLNotifyReceiverColumn(session, table); err != nil {
		t.Fatal(err)
	}
	rows, err := session.Query("SELECT notify_receiver_id FROM `" + table + "` ORDER BY id")
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"", "8", "0"} {
		if len(rows) != 3 || string(rows[i]["notify_receiver_id"]) != want {
			t.Fatalf("receiver values not preserved: %v", rows)
		}
	}
	target := `{"format":"webhook_target","version":1,"group_ids":[8],"template_id":5}`
	if _, err := session.Exec("UPDATE `"+table+"` SET notify_receiver_id = ? WHERE id = 2", target); err != nil {
		t.Fatal(err)
	}
	if err := ensureMySQLNotifyReceiverColumn(session, table); err != nil {
		t.Fatal(err)
	}
	rows, err = session.Query("SELECT notify_receiver_id FROM `" + table + "` WHERE id = 2")
	if err != nil || len(rows) != 1 || string(rows[0]["notify_receiver_id"]) != target {
		t.Fatalf("JSON round trip after repeated migration: %v, %v", rows, err)
	}
	columnType, nullable, err := mysqlColumnDefinition(session, table, "notify_receiver_id")
	if err != nil || columnType != "varchar" || nullable {
		t.Fatalf("unexpected column: %s nullable=%v err=%v", columnType, nullable, err)
	}
}
