package controllers

import (
	"fmt"
	"testing"
)

func TestNormalizePage(t *testing.T) {
	tests := []struct {
		page       int
		totalPages int
		want       int
	}{
		{0, 5, 1},        // zero page -> 1
		{-1, 5, 1},       // negative page -> 1
		{-100, 10, 1},    // large negative -> 1
		{1, 5, 1},        // first page stays 1
		{3, 5, 3},        // middle page stays 3
		{5, 5, 5},        // last page stays 5
		{6, 5, 5},        // beyond last -> last
		{1000, 5, 5},     // way beyond -> last
		{0, 1, 1},        // zero page, single page -> 1
		{1, 1, 1},        // page 1 of 1 stays 1
		{2, 1, 1},        // page 2 of 1 -> 1
		{0, 0, 1},        // totalPages 0 treated as 1, page 0 -> 1
		{-5, 0, 1},       // negative page, totalPages 0 -> 1
		{100, 0, 1},      // large page, totalPages 0 -> 1
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("page=%d/total=%d", tt.page, tt.totalPages), func(t *testing.T) {
			got := NormalizePage(tt.page, tt.totalPages)
			if got != tt.want {
				t.Errorf("NormalizePage(%d, %d) = %d, want %d", tt.page, tt.totalPages, got, tt.want)
			}
		})
	}
}

func TestBuildPagination(t *testing.T) {
	// Basic case: 5 pages, current page 3
	items := BuildPagination(3, 5)
	if len(items) != 5 {
		t.Fatalf("BuildPagination(3, 5): got %d items, want 5", len(items))
	}
	for i, item := range items {
		expectedNum := i + 1
		if item.Num != expectedNum {
			t.Errorf("item[%d].Num = %d, want %d", i, item.Num, expectedNum)
		}
		expectedURL := fmt.Sprintf("/listArticle?page=%d", expectedNum)
		if item.URL != expectedURL {
			t.Errorf("item[%d].URL = %q, want %q", i, item.URL, expectedURL)
		}
		if expectedNum == 3 {
			if item.Active != "active" {
				t.Errorf("item[%d].Active = %q, want \"active\"", i, item.Active)
			}
		} else {
			if item.Active != "" {
				t.Errorf("item[%d].Active = %q, want empty", i, item.Active)
			}
		}
	}

	// Single page
	items = BuildPagination(1, 1)
	if len(items) != 1 {
		t.Fatalf("BuildPagination(1, 1): got %d items, want 1", len(items))
	}
	if items[0].Active != "active" {
		t.Errorf("single page: Active = %q, want \"active\"", items[0].Active)
	}
	if items[0].URL != "/listArticle?page=1" {
		t.Errorf("single page: URL = %q, want \"/listArticle?page=1\"", items[0].URL)
	}

	// Zero totalPages treated as 1
	items = BuildPagination(1, 0)
	if len(items) != 1 {
		t.Fatalf("BuildPagination(1, 0): got %d items, want 1", len(items))
	}

	// Current page at first boundary
	items = BuildPagination(1, 3)
	if items[0].Active != "active" {
		t.Errorf("first page: items[0].Active = %q, want \"active\"", items[0].Active)
	}
	if items[1].Active != "" {
		t.Errorf("first page: items[1].Active = %q, want empty", items[1].Active)
	}

	// Current page at last boundary
	items = BuildPagination(3, 3)
	if items[2].Active != "active" {
		t.Errorf("last page: items[2].Active = %q, want \"active\"", items[2].Active)
	}
	if items[0].Active != "" {
		t.Errorf("last page: items[0].Active = %q, want empty", items[0].Active)
	}
}

func TestBuildPaginationURLs(t *testing.T) {
	// Verify all URLs are complete, directly usable links (no JS needed)
	items := BuildPagination(2, 4)
	for i, item := range items {
		expected := fmt.Sprintf("/listArticle?page=%d", i+1)
		if item.URL != expected {
			t.Errorf("item[%d].URL = %q, want %q", i, item.URL, expected)
		}
	}
}
