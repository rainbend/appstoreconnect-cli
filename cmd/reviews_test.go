package cmd

import (
	"testing"
	"time"
)

func TestParseRatings(t *testing.T) {
	ratings, err := parseRatings("1, 3,5")
	if err != nil {
		t.Fatalf("parse ratings: %v", err)
	}

	want := []int{1, 3, 5}
	if len(ratings) != len(want) {
		t.Fatalf("ratings length = %d, want %d", len(ratings), len(want))
	}
	for i := range want {
		if ratings[i] != want[i] {
			t.Fatalf("ratings[%d] = %d, want %d", i, ratings[i], want[i])
		}
	}
}

func TestParseRatingsRejectsInvalidValues(t *testing.T) {
	if _, err := parseRatings("0,6"); err == nil {
		t.Fatal("expected invalid rating error")
	}
}

func TestParsePublishedResponseFilter(t *testing.T) {
	tests := []struct {
		value   string
		wantNil bool
		want    bool
	}{
		{value: "all", wantNil: true},
		{value: "responded", want: true},
		{value: "replied", want: true},
		{value: "unresponded", want: false},
		{value: "unreplied", want: false},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, err := parsePublishedResponseFilter(test.value)
			if err != nil {
				t.Fatalf("parse filter: %v", err)
			}
			if test.wantNil {
				if got != nil {
					t.Fatalf("filter = %v, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatal("filter = nil")
			}
			if *got != test.want {
				t.Fatalf("filter = %t, want %t", *got, test.want)
			}
		})
	}
}

func TestParseReviewCreatedSinceDays(t *testing.T) {
	now := time.Date(2026, 6, 8, 12, 30, 0, 0, time.UTC)
	got, err := parseReviewCreatedSince("", 7, now)
	if err != nil {
		t.Fatalf("parse created since: %v", err)
	}

	want := now.AddDate(0, 0, -7)
	if got == nil || !got.Equal(want) {
		t.Fatalf("created since = %v, want %v", got, want)
	}
}

func TestParseReviewCreatedSinceDate(t *testing.T) {
	got, err := parseReviewCreatedSince("2026-06-01", 0, time.Now())
	if err != nil {
		t.Fatalf("parse created since: %v", err)
	}
	if got == nil {
		t.Fatal("created since = nil")
	}
	if got.Year() != 2026 || got.Month() != 6 || got.Day() != 1 {
		t.Fatalf("created since date = %v, want 2026-06-01", got)
	}
	if got.Hour() != 0 || got.Minute() != 0 || got.Second() != 0 {
		t.Fatalf("created since time = %v, want midnight", got)
	}
}

func TestParseReviewCreatedSinceRFC3339(t *testing.T) {
	got, err := parseReviewCreatedSince("2026-06-01T08:15:30Z", 0, time.Now())
	if err != nil {
		t.Fatalf("parse created since: %v", err)
	}

	want := time.Date(2026, 6, 1, 8, 15, 30, 0, time.UTC)
	if got == nil || !got.Equal(want) {
		t.Fatalf("created since = %v, want %v", got, want)
	}
}

func TestParseReviewCreatedSinceRejectsConflicts(t *testing.T) {
	if _, err := parseReviewCreatedSince("2026-06-01", 7, time.Now()); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestParseReviewCreatedSinceRejectsNegativeDays(t *testing.T) {
	if _, err := parseReviewCreatedSince("", -1, time.Now()); err == nil {
		t.Fatal("expected negative days error")
	}
}
