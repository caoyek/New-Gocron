package models

import (
	"testing"
	"time"
)

func TestTaskOptionsValidation(t *testing.T) {
	valid := Task{Protocol: TaskHTTP, HttpMethod: TaskHttpMethodPost, HttpBodyMode: "json", HttpBody: `{"ok":true}`, HttpHeaders: `{"Authorization":"Bearer example"}`, LogRetentionDays: 90}
	if err := valid.ValidateOptions(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		change func(*Task)
	}{
		{"invalid headers", func(v *Task) { v.HttpHeaders = `{"X-Test":123}` }},
		{"injection", func(v *Task) { v.HttpHeaders = `{"X-Test":"a\r\nInjected: yes"}` }},
		{"invalid name", func(v *Task) { v.HttpHeaders = `{"Bad Name":"value"}` }},
		{"null", func(v *Task) { v.HttpHeaders = `null` }},
		{"null value", func(v *Task) { v.HttpHeaders = `{"X-Test":null}` }},
		{"reserved", func(v *Task) { v.HttpHeaders = `{"Content-Length":"100"}` }},
		{"duplicate headers", func(v *Task) { v.HttpHeaders = `{"Authorization":"first","authorization":"second"}` }},
		{"body", func(v *Task) { v.HttpBody = `{` }},
		{"get body", func(v *Task) { v.HttpMethod = TaskHTTPMethodGet }},
		{"mode", func(v *Task) { v.HttpBodyMode = "xml" }},
		{"expiry", func(v *Task) { v.ExpiresAt = "2026-02-30 00:00:00" }},
		{"retention", func(v *Task) { v.LogRetentionDays = -1 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			v := valid
			test.change(&v)
			if v.ValidateOptions() == nil {
				t.Fatal("invalid option accepted")
			}
		})
	}
	if err := (Task{}).ValidateOptions(); err != nil {
		t.Fatal("legacy task", err)
	}
	for _, days := range []int{0, 7, 30, 90, 180, 365} {
		if err := (Task{LogRetentionDays: days}).ValidateOptions(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestTaskExpiration(t *testing.T) {
	boundary := time.Date(2026, 10, 9, 12, 0, 0, 0, time.Local)
	task := Task{ExpiresAt: boundary.Format(DefaultTimeFormat)}
	if task.Expired(boundary.Add(-time.Second)) || !task.Expired(boundary) || !task.Expired(boundary.Add(time.Second)) {
		t.Fatal("incorrect expiration boundary")
	}
	if (Task{}).Expired(boundary) {
		t.Fatal("legacy task must not expire")
	}
}
