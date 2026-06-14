package models

import (
	"testing"
)

func TestPublicBlogBaseFilters(t *testing.T) {
	filters := PublicBlogBaseFilters()
	if len(filters) != 2 {
		t.Fatalf("expected 2 filter elements, got %d", len(filters))
	}
	if filters[0] != "status" {
		t.Errorf("expected first filter key to be 'status', got %v", filters[0])
	}
	if filters[1] != "public" {
		t.Errorf("expected first filter value to be 'public', got %v", filters[1])
	}
}

func TestMergeFilters(t *testing.T) {
	a := []interface{}{"status", "public"}
	b := []interface{}{"type", "original"}
	c := []interface{}{"catalogid", "1"}

	merged := MergeFilters(a, b, c)
	if len(merged) != 6 {
		t.Fatalf("expected 6 elements, got %d", len(merged))
	}
	expected := []interface{}{"status", "public", "type", "original", "catalogid", "1"}
	for i, v := range expected {
		if merged[i] != v {
			t.Errorf("merged[%d] = %v, want %v", i, merged[i], v)
		}
	}
}

func TestMergeFiltersEmpty(t *testing.T) {
	merged := MergeFilters()
	if len(merged) != 0 {
		t.Fatalf("expected 0 elements for empty merge, got %d", len(merged))
	}
}

func TestMergeFiltersNil(t *testing.T) {
	a := []interface{}{"status", "public"}
	merged := MergeFilters(a, nil)
	if len(merged) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(merged))
	}
}
