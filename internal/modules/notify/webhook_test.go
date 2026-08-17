package notify

import (
	"testing"

	"github.com/caoyek/New-Gocron/internal/models"
)

func TestWebHookGetActiveGroups(t *testing.T) {
	setting := models.WebHook{Groups: []models.WebHookGroup{
		{Id: 10, Name: "运维群", Url: "https://example.com/10"},
		{Id: 20, Name: "业务群", Url: "https://example.com/20"},
		{Id: 30, Name: "告警群", Url: "https://example.com/30"},
	}}
	webHook := new(WebHook)

	groups := webHook.getActiveGroups(setting, Message{"task_receiver_id": "30,10"})
	if len(groups) != 2 || groups[0].Id != 10 || groups[1].Id != 30 {
		t.Fatalf("unexpected selected groups: %+v", groups)
	}
}

func TestWebHookGetActiveGroupsFromVersionedTarget(t *testing.T) {
	setting := models.WebHook{Groups: []models.WebHookGroup{
		{Id: 10, Name: "运维群", Url: "https://example.com/10"},
		{Id: 20, Name: "业务群", Url: "https://example.com/20"},
		{Id: 30, Name: "告警群", Url: "https://example.com/30"},
	}}

	groups := new(WebHook).getActiveGroups(setting, Message{
		"task_receiver_id": `{"format":"webhook_target","version":1,"group_ids":[30,10],"template_id":2}`,
	})
	if len(groups) != 2 || groups[0].Id != 10 || groups[1].Id != 30 {
		t.Fatalf("unexpected selected groups: %+v", groups)
	}
}

func TestWebHookGetActiveGroupsFallsBackToFirstGroup(t *testing.T) {
	setting := models.WebHook{Groups: []models.WebHookGroup{
		{Id: 10, Name: "默认企微群", Url: "https://example.com/10"},
		{Id: 20, Name: "业务群", Url: "https://example.com/20"},
	}}

	groups := new(WebHook).getActiveGroups(setting, Message{"task_receiver_id": ""})
	if len(groups) != 1 || groups[0].Id != 10 {
		t.Fatalf("unexpected fallback group: %+v", groups)
	}
}

func TestWebHookGetActiveGroupsHandlesEmptyConfiguration(t *testing.T) {
	groups := new(WebHook).getActiveGroups(models.WebHook{}, Message{"task_receiver_id": ""})
	if len(groups) != 0 {
		t.Fatalf("expected no groups, got %+v", groups)
	}
}

func TestWebHookGetActiveTemplateFromVersionedTarget(t *testing.T) {
	setting := models.WebHook{Templates: []models.WebHookTemplate{
		{Id: 1, Name: "默认模板", Content: "default", IsDefault: true},
		{Id: 2, Name: "失败告警", Content: "failed"},
	}}

	template, ok := new(WebHook).getActiveTemplate(setting, Message{
		"task_receiver_id": `{"format":"webhook_target","version":1,"group_ids":[10],"template_id":2}`,
	})
	if !ok || template.Id != 2 || template.Content != "failed" {
		t.Fatalf("unexpected selected template: %+v", template)
	}
}

func TestWebHookGetActiveTemplateFallsBackForLegacyAndMissingSelection(t *testing.T) {
	setting := models.WebHook{Templates: []models.WebHookTemplate{
		{Id: 1, Name: "默认模板", Content: "default", IsDefault: true},
		{Id: 2, Name: "失败告警", Content: "failed"},
	}}
	webHook := new(WebHook)

	legacy, ok := webHook.getActiveTemplate(setting, Message{"task_receiver_id": "10,20"})
	if !ok || legacy.Id != 1 {
		t.Fatalf("unexpected legacy fallback: %+v", legacy)
	}

	missing, ok := webHook.getActiveTemplate(setting, Message{
		"task_receiver_id": `{"format":"webhook_target","version":1,"group_ids":[10],"template_id":999}`,
	})
	if !ok || missing.Id != 1 {
		t.Fatalf("unexpected missing-template fallback: %+v", missing)
	}
}
