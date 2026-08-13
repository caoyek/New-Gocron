package task

import (
	"reflect"
	"testing"
)

func TestParsePinnedIds(t *testing.T) {
	got := parsePinnedIds("9, 3,invalid,-1,9,7")
	want := []int{9, 3, 7}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
