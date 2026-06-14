package controllers

import "testing"

// normalizePage must keep every requested page inside [1, totalPages] so that
// a missing, zero, negative or oversized "page" param never produces a
// negative offset or an empty first page.
func TestNormalizePage(t *testing.T) {
	cases := []struct {
		name       string
		page       int
		totalPages int
		want       int
	}{
		{"zero falls back to first", 0, 5, 1},
		{"negative falls back to first", -3, 5, 1},
		{"first stays first", 1, 5, 1},
		{"in range is kept", 3, 5, 3},
		{"last stays last", 5, 5, 5},
		{"over range clamps to last", 9, 5, 5},
		{"no data still resolves to first", 1, 0, 1},
		{"over range with no data clamps to first", 7, 0, 1},
		{"negative with no data clamps to first", -1, 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizePage(tc.page, tc.totalPages); got != tc.want {
				t.Fatalf("normalizePage(%d, %d) = %d, want %d", tc.page, tc.totalPages, got, tc.want)
			}
		})
	}
}

func TestBuildPaginationSinglePage(t *testing.T) {
	p := buildPagination(1, 1)
	if p.HasPrev {
		t.Errorf("single page should not have a previous link")
	}
	if p.HasNext {
		t.Errorf("single page should not have a next link")
	}
	if p.PrevURL != "" || p.NextURL != "" {
		t.Errorf("single page prev/next URLs should be empty, got prev=%q next=%q", p.PrevURL, p.NextURL)
	}
	if len(p.Items) != 1 {
		t.Fatalf("expected 1 page item, got %d", len(p.Items))
	}
	if !p.Items[0].Active {
		t.Errorf("the only page should be marked active")
	}
	if p.Items[0].URL != "/listArticle?page=1" {
		t.Errorf("unexpected item URL: %q", p.Items[0].URL)
	}
}

func TestBuildPaginationFirstPage(t *testing.T) {
	p := buildPagination(1, 3)
	if p.HasPrev {
		t.Errorf("first page should not have a previous link")
	}
	if !p.HasNext {
		t.Errorf("first page of many should have a next link")
	}
	if p.NextURL != "/listArticle?page=2" {
		t.Errorf("next URL = %q, want /listArticle?page=2", p.NextURL)
	}
	if !p.Items[0].Active {
		t.Errorf("page 1 item should be active")
	}
	if p.Items[1].Active || p.Items[2].Active {
		t.Errorf("non-current page items must not be active")
	}
}

func TestBuildPaginationMiddlePage(t *testing.T) {
	p := buildPagination(2, 3)
	if !p.HasPrev || !p.HasNext {
		t.Errorf("a middle page should have both previous and next links")
	}
	if p.PrevURL != "/listArticle?page=1" {
		t.Errorf("prev URL = %q, want /listArticle?page=1", p.PrevURL)
	}
	if p.NextURL != "/listArticle?page=3" {
		t.Errorf("next URL = %q, want /listArticle?page=3", p.NextURL)
	}
	if !p.Items[1].Active {
		t.Errorf("page 2 item should be active")
	}
}

func TestBuildPaginationLastPage(t *testing.T) {
	p := buildPagination(3, 3)
	if !p.HasPrev {
		t.Errorf("last page should have a previous link")
	}
	if p.HasNext {
		t.Errorf("last page should not have a next link")
	}
	if p.PrevURL != "/listArticle?page=2" {
		t.Errorf("prev URL = %q, want /listArticle?page=2", p.PrevURL)
	}
	if p.NextURL != "" {
		t.Errorf("last page next URL should be empty, got %q", p.NextURL)
	}
}

// An out-of-range current page must be clamped so the boundary state (and the
// active highlight) still lands on a real page.
func TestBuildPaginationClampsOutOfRangeCurrent(t *testing.T) {
	p := buildPagination(99, 3)
	if p.Current != 3 {
		t.Errorf("current = %d, want clamped to 3", p.Current)
	}
	if p.HasNext {
		t.Errorf("clamped-to-last page should not have a next link")
	}
	if !p.Items[2].Active {
		t.Errorf("last page item should be active after clamping")
	}
}

func TestBuildPaginationNoData(t *testing.T) {
	p := buildPagination(1, 0)
	if p.TotalPages != 1 {
		t.Errorf("empty list should still expose a single page, got %d", p.TotalPages)
	}
	if len(p.Items) != 1 {
		t.Fatalf("empty list should render exactly one page item, got %d", len(p.Items))
	}
	if p.HasPrev || p.HasNext {
		t.Errorf("empty list should have no prev/next links")
	}
}

func TestArticleActionURL(t *testing.T) {
	cases := []struct {
		action string
		id     int
		want   string
	}{
		{"look", 7, "/lookArticle?id=7"},
		{"update", 42, "/updateArticle?id=42"},
		{"delete", 5, "/deleteArticle?id=5"},
		{"unknown", 1, "/listArticle?id=1"},
	}
	for _, tc := range cases {
		if got := articleActionURL(tc.action, tc.id); got != tc.want {
			t.Errorf("articleActionURL(%q, %d) = %q, want %q", tc.action, tc.id, got, tc.want)
		}
	}
}

func TestArticlePageURL(t *testing.T) {
	if got := articlePageURL(2); got != "/listArticle?page=2" {
		t.Errorf("articlePageURL(2) = %q, want /listArticle?page=2", got)
	}
}
