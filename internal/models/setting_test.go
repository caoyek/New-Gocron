package models

import "testing"

func TestFormatWebhookIncludesLegacyAndNamedGroups(t *testing.T) {
	settings := []Setting{
		{Id: 1, Code: WebhookCode, Key: WebhookTemplateKey, Value: "template"},
		{Id: 2, Code: WebhookCode, Key: WebhookUrlKey, Value: "https://example.com/default"},
		{Id: 3, Code: WebhookCode, Key: WebhookGroupKey, Value: `{"name":"业务群","url":"https://example.com/business"}`},
		{Id: 4, Code: WebhookCode, Key: WebhookTemplateKeyPrefix + "失败告警", Value: "failed template"},
	}
	webHook := WebHook{Groups: make([]WebHookGroup, 0)}

	if err := new(Setting).formatWebhook(settings, &webHook); err != nil {
		t.Fatal(err)
	}
	if webHook.Template != "template" {
		t.Fatalf("unexpected template: %s", webHook.Template)
	}
	if len(webHook.Templates) != 2 {
		t.Fatalf("expected two templates, got %+v", webHook.Templates)
	}
	if webHook.Templates[0].Id != 1 || webHook.Templates[0].Name != "默认模板" || !webHook.Templates[0].IsDefault {
		t.Fatalf("unexpected default template: %+v", webHook.Templates[0])
	}
	if webHook.Templates[1].Id != 4 || webHook.Templates[1].Name != "失败告警" || webHook.Templates[1].Content != "failed template" {
		t.Fatalf("unexpected named template: %+v", webHook.Templates[1])
	}
	if len(webHook.Groups) != 2 {
		t.Fatalf("expected two groups, got %+v", webHook.Groups)
	}
	if webHook.Groups[0].Id != 2 || webHook.Groups[0].Name != "默认企微群" {
		t.Fatalf("unexpected legacy group: %+v", webHook.Groups[0])
	}
	if webHook.Groups[1].Id != 3 || webHook.Groups[1].Name != "业务群" {
		t.Fatalf("unexpected named group: %+v", webHook.Groups[1])
	}
}

func TestFormatWebhookRejectsInvalidGroup(t *testing.T) {
	settings := []Setting{{Id: 1, Code: WebhookCode, Key: WebhookGroupKey, Value: "{"}}
	webHook := WebHook{Groups: make([]WebHookGroup, 0)}

	if err := new(Setting).formatWebhook(settings, &webHook); err == nil {
		t.Fatal("expected invalid group configuration to fail")
	}
}

func TestFormatWebhookKeepsDefaultTemplateFirst(t *testing.T) {
	settings := []Setting{
		{Id: 1, Code: WebhookCode, Key: WebhookTemplateKeyPrefix + "失败告警", Value: "failed"},
		{Id: 2, Code: WebhookCode, Key: WebhookTemplateKey, Value: "default"},
	}
	webHook := WebHook{Templates: make([]WebHookTemplate, 0)}

	if err := new(Setting).formatWebhook(settings, &webHook); err != nil {
		t.Fatal(err)
	}
	if len(webHook.Templates) != 2 || !webHook.Templates[0].IsDefault || webHook.Templates[0].Content != "default" {
		t.Fatalf("expected default template first, got %+v", webHook.Templates)
	}
}

func TestDecodeWebhookTarget(t *testing.T) {
	target, err := DecodeWebhookTarget(`{"format":"webhook_target","version":1,"group_ids":[10,20],"template_id":30}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(target.GroupIds) != 2 || target.GroupIds[0] != 10 || target.TemplateId != 30 {
		t.Fatalf("unexpected target: %+v", target)
	}

	if _, err = DecodeWebhookTarget(`{"format":"other","version":1}`); err == nil {
		t.Fatal("expected unsupported target format to fail")
	}
	if _, err = DecodeWebhookTarget("10,20"); err == nil {
		t.Fatal("expected legacy receiver IDs not to decode as a versioned target")
	}
}

func TestWebhookValidateTarget(t *testing.T) {
	webHook := WebHook{
		Groups:    []WebHookGroup{{Id: 10}, {Id: 20}},
		Templates: []WebHookTemplate{{Id: 30}, {Id: 40}},
	}

	tests := []struct {
		name    string
		target  WebhookTarget
		wantErr bool
	}{
		{name: "valid", target: WebhookTarget{GroupIds: []int{10, 20}, TemplateId: 30}},
		{name: "missing groups", target: WebhookTarget{TemplateId: 30}, wantErr: true},
		{name: "unknown group", target: WebhookTarget{GroupIds: []int{99}, TemplateId: 30}, wantErr: true},
		{name: "missing template", target: WebhookTarget{GroupIds: []int{10}}, wantErr: true},
		{name: "unknown template", target: WebhookTarget{GroupIds: []int{10}, TemplateId: 99}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := webHook.ValidateTarget(test.target)
			if (err != nil) != test.wantErr {
				t.Fatalf("ValidateTarget() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
