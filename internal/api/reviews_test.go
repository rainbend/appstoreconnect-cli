package api

import (
	"testing"
	"time"
)

func TestCustomerReviewsQueryValues(t *testing.T) {
	publishedResponse := true
	values := customerReviewsQueryValues(CustomerReviewsQuery{
		Territories:       []string{"USA", "CHN"},
		Ratings:           []int{1, 5},
		PublishedResponse: &publishedResponse,
		IncludeResponse:   true,
		Sort:              "-createdDate",
	}, 50)

	tests := map[string]string{
		"filter[territory]":               "USA,CHN",
		"filter[rating]":                  "1,5",
		"exists[publishedResponse]":       "true",
		"include":                         "response",
		"sort":                            "-createdDate",
		"limit":                           "50",
		"fields[customerReviewResponses]": "responseBody,lastModifiedDate,state,review",
	}

	for key, want := range tests {
		if got := values.Get(key); got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}

func TestMapResponsesByReviewID(t *testing.T) {
	reviews := []CustomerReview{
		{
			ID: "review-1",
			Relationships: CustomerReviewRelationships{
				Response: &OptionalRelationshipData{
					Data: &ResourceIdentifier{Type: "customerReviewResponses", ID: "response-1"},
				},
			},
		},
		{ID: "review-2"},
	}
	included := []CustomerReviewResponseV1{
		{
			ID: "response-1",
			Attributes: CustomerReviewResponseAttributes{
				ResponseBody: "Thanks",
			},
		},
		{
			ID: "response-2",
			Relationships: &CustomerReviewResponseRelationships{
				Review: RelationshipData{
					Data: ResourceIdentifier{Type: "customerReviews", ID: "review-2"},
				},
			},
		},
	}

	responses := mapResponsesByReviewID(reviews, included)
	if got := responses["review-1"]; got == nil || got.ID != "response-1" {
		t.Fatalf("review-1 response = %#v, want response-1", got)
	}
	if got := responses["review-2"]; got == nil || got.ID != "response-2" {
		t.Fatalf("review-2 response = %#v, want response-2", got)
	}
}

func TestCustomerReviewCreatedAt(t *testing.T) {
	got, err := customerReviewCreatedAt(CustomerReview{
		ID: "review-1",
		Attributes: CustomerReviewAttributes{
			CreatedDate: "2026-06-01T08:15:30-07:00",
		},
	})
	if err != nil {
		t.Fatalf("parse created date: %v", err)
	}

	want := time.Date(2026, 6, 1, 15, 15, 30, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("created date = %v, want %v", got, want)
	}
}
