package models

import "testing"

func TestPinnedTaskOrder(t *testing.T) {
	got := pinnedTaskOrder([]int{9, 3})
	want := "CASE t.id WHEN 9 THEN 0 WHEN 3 THEN 1 ELSE 2 END ASC"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestWebhookReceiverReferencesGroup(t *testing.T) {
	tests := []struct {
		name           string
		receiverIds    string
		groupId        int
		defaultGroupId int
		want           bool
	}{
		{name: "versioned target", receiverIds: `{"format":"webhook_target","version":1,"group_ids":[10,20],"template_id":30}`, groupId: 20, defaultGroupId: 10, want: true},
		{name: "versioned target not selected", receiverIds: `{"format":"webhook_target","version":1,"group_ids":[10],"template_id":30}`, groupId: 20, defaultGroupId: 10},
		{name: "legacy comma IDs", receiverIds: "10, 20", groupId: 20, defaultGroupId: 10, want: true},
		{name: "legacy empty uses default", receiverIds: "", groupId: 10, defaultGroupId: 10, want: true},
		{name: "legacy empty ignores non-default", receiverIds: "", groupId: 20, defaultGroupId: 10},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := webhookReceiverReferencesGroup(test.receiverIds, test.groupId, test.defaultGroupId)
			if got != test.want {
				t.Fatalf("webhookReceiverReferencesGroup() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWebhookReceiverReferencesTemplate(t *testing.T) {
	if !webhookReceiverReferencesTemplate(`{"format":"webhook_target","version":1,"group_ids":[10],"template_id":30}`, 30) {
		t.Fatal("expected versioned target to reference template")
	}
	if webhookReceiverReferencesTemplate(`{"format":"webhook_target","version":1,"group_ids":[10],"template_id":30}`, 40) {
		t.Fatal("unexpected template reference")
	}
	if webhookReceiverReferencesTemplate("10,20", 30) {
		t.Fatal("legacy receiver IDs must not reference a named template")
	}
}
