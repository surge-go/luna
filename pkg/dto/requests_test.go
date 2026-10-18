package dto

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

func TestIdsReqJSON(t *testing.T) {
	var req IdsReq
	if err := json.Unmarshal([]byte(`{"ids":[1,2,3]}`), &req); err != nil {
		t.Fatalf("unmarshal numeric IDs: %v", err)
	}
	if len(req.Ids) != 3 || req.Ids[0] != 1 || req.Ids[1] != 2 || req.Ids[2] != 3 {
		t.Fatalf("Ids = %v, want [1 2 3]", req.Ids)
	}

	if err := json.Unmarshal([]byte(`{"ids":["1"]}`), &req); err == nil {
		t.Fatal("unmarshal string ID error = nil, want an error")
	}
}

func TestSortReqDesc(t *testing.T) {
	desc := 0
	asc := 1

	tests := []struct {
		name string
		req  SortReq
		want bool
	}{
		{name: "order omitted defaults to desc", want: true},
		{name: "explicit desc", req: SortReq{Order: &desc}, want: true},
		{name: "explicit asc", req: SortReq{Order: &asc}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.Desc(); got != tt.want {
				t.Fatalf("Desc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageReqPagination(t *testing.T) {
	tests := []struct {
		name       string
		req        PageReq
		wantOffset int
		wantLimit  int
	}{
		{
			name:       "defaults",
			wantOffset: 0,
			wantLimit:  1,
		},
		{
			name:       "values below one are normalized",
			req:        PageReq{Page: -1, Size: -1},
			wantOffset: 0,
			wantLimit:  1,
		},
		{
			name:       "requested page and size",
			req:        PageReq{Page: 3, Size: 25},
			wantOffset: 50,
			wantLimit:  25,
		},
		{
			name:       "minimum size participates in offset",
			req:        PageReq{Page: 3},
			wantOffset: 2,
			wantLimit:  1,
		},
		{
			name:       "maximum page size",
			req:        PageReq{Page: 2, Size: MaxPageSize},
			wantOffset: MaxPageSize,
			wantLimit:  MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.req.Offset(); got != tt.wantOffset {
				t.Fatalf("Offset() = %d, want %d", got, tt.wantOffset)
			}

			if got := tt.req.Limit(); got != tt.wantLimit {
				t.Fatalf("Limit() = %d, want %d", got, tt.wantLimit)
			}
		})
	}
}

func TestPageReqClampsOversizedLimit(t *testing.T) {
	req := PageReq{Page: 2, Size: MaxPageSize + 1}

	if got := req.Limit(); got != MaxPageSize {
		t.Fatalf("Limit() = %d, want %d", got, MaxPageSize)
	}
	if got := req.Offset(); got != MaxPageSize {
		t.Fatalf("Offset() = %d, want %d", got, MaxPageSize)
	}
}

func TestPageReqClampsOffsetOverflow(t *testing.T) {
	req := PageReq{Page: math.MaxInt, Size: MaxPageSize}

	if got := req.Offset(); got != math.MaxInt {
		t.Fatalf("Offset() = %d, want %d", got, math.MaxInt)
	}
}

func TestTimeRangeReqParse(t *testing.T) {
	tests := []struct {
		name     string
		req      TimeRangeReq
		wantFrom *time.Time
		wantTo   *time.Time
		wantErr  bool
	}{
		{
			name: "empty range",
		},
		{
			name:     "from only",
			req:      TimeRangeReq{From: "2026-08-15 09:30:00"},
			wantFrom: timePointer(t, "2026-08-15 09:30:00", DefaultTimezone),
		},
		{
			name:   "to only",
			req:    TimeRangeReq{To: "2026-08-15 18:00:00"},
			wantTo: timePointer(t, "2026-08-15 18:00:00", DefaultTimezone),
		},
		{
			name:     "closed range",
			req:      TimeRangeReq{From: "2026-08-15 09:30:00", To: "2026-08-15 18:00:00"},
			wantFrom: timePointer(t, "2026-08-15 09:30:00", DefaultTimezone),
			wantTo:   timePointer(t, "2026-08-15 18:00:00", DefaultTimezone),
		},
		{
			name:     "equal boundaries",
			req:      TimeRangeReq{From: "2026-08-15 09:30:00", To: "2026-08-15 09:30:00"},
			wantFrom: timePointer(t, "2026-08-15 09:30:00", DefaultTimezone),
			wantTo:   timePointer(t, "2026-08-15 09:30:00", DefaultTimezone),
		},
		{
			name:     "explicit UTC timezone",
			req:      TimeRangeReq{From: "2026-08-15 09:30:00", Timezone: "UTC"},
			wantFrom: timePointer(t, "2026-08-15 09:30:00", "UTC"),
		},
		{
			name:     "blank timezone uses default",
			req:      TimeRangeReq{From: "2026-08-15 09:30:00", Timezone: "  "},
			wantFrom: timePointer(t, "2026-08-15 09:30:00", DefaultTimezone),
		},
		{
			name:    "invalid from format",
			req:     TimeRangeReq{From: "2026-08-15"},
			wantErr: true,
		},
		{
			name:    "invalid calendar date",
			req:     TimeRangeReq{To: "2026-02-30 12:00:00"},
			wantErr: true,
		},
		{
			name:    "from after to",
			req:     TimeRangeReq{From: "2026-08-16 00:00:00", To: "2026-08-15 23:59:59"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			from, to, err := tt.req.Parse()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Parse() error = nil, want an error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			assertOptionalTimeEqual(t, "from", from, tt.wantFrom)
			assertOptionalTimeEqual(t, "to", to, tt.wantTo)
		})
	}
}

func TestTimeRangeReqRejectsInvalidTimezone(t *testing.T) {
	req := TimeRangeReq{Timezone: "Invalid/Timezone"}

	if _, _, err := req.Parse(); !errors.Is(err, ErrInvalidTimezone) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrInvalidTimezone)
	}
}

func timePointer(t *testing.T, value, timezone string) *time.Time {
	t.Helper()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		t.Fatalf("load test timezone %q: %v", timezone, err)
	}
	parsed, err := time.ParseInLocation(DateTimeLayout, value, location)
	if err != nil {
		t.Fatalf("parse test time %q: %v", value, err)
	}
	return &parsed
}

func assertOptionalTimeEqual(t *testing.T, name string, got, want *time.Time) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
		return
	}
	if !got.Equal(*want) {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}
