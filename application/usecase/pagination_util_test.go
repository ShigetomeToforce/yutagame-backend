package usecase

import "testing"

func TestCalculatePagination(t *testing.T) {
	tests := []struct {
		name       string
		totalCount int64
		page       int
		limit      int
		want       PaginationParam
	}{
		{name: "first page", totalCount: 95, page: 1, limit: 20, want: PaginationParam{Offset: 0, TotalPages: 5, ActivePage: 1}},
		{name: "last partial page", totalCount: 95, page: 5, limit: 20, want: PaginationParam{Offset: 80, TotalPages: 5, ActivePage: 5}},
		{name: "empty result still has page one", totalCount: 0, page: 1, limit: 20, want: PaginationParam{Offset: 0, TotalPages: 1, ActivePage: 1}},
		{name: "page below range", totalCount: 40, page: 0, limit: 20, want: PaginationParam{Offset: 0, TotalPages: 2, ActivePage: 1}},
		{name: "page above range", totalCount: 40, page: 9, limit: 20, want: PaginationParam{Offset: 20, TotalPages: 2, ActivePage: 2}},
		{name: "invalid limit", totalCount: 2, page: 1, limit: 0, want: PaginationParam{Offset: 0, TotalPages: 2, ActivePage: 1}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CalculatePagination(test.totalCount, test.page, test.limit); got != test.want {
				t.Fatalf("CalculatePagination() = %+v, want %+v", got, test.want)
			}
		})
	}
}
