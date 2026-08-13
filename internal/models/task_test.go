package models

import "testing"

func TestPinnedTaskOrder(t *testing.T) {
	got := pinnedTaskOrder([]int{9, 3})
	want := "CASE t.id WHEN 9 THEN 0 WHEN 3 THEN 1 ELSE 2 END ASC"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
