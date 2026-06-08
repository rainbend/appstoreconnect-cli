package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const customerReviewsPageLimit = 200

func (c *Client) ListCustomerReviews(appID string, query CustomerReviewsQuery) (*CustomerReviewsResult, error) {
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Sort == "" {
		query.Sort = "-createdDate"
	}
	if query.CreatedSince != nil && query.Sort != "-createdDate" {
		return nil, fmt.Errorf("created date filtering requires sort -createdDate")
	}

	remaining := query.Limit
	path := fmt.Sprintf("/apps/%s/customerReviews?%s", appID, customerReviewsQueryValues(query, minInt(remaining, customerReviewsPageLimit)).Encode())

	var reviews []CustomerReview
	var included []CustomerReviewResponseV1

	for path != "" && remaining > 0 {
		data, err := c.do("GET", path, nil)
		if err != nil {
			return nil, err
		}

		var resp CustomerReviewsResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, fmt.Errorf("decoding response: %w", err)
		}

		if len(resp.Data) == 0 {
			break
		}

		foundOlderReview := false
		for _, review := range resp.Data {
			if query.CreatedSince != nil {
				createdAt, err := customerReviewCreatedAt(review)
				if err != nil {
					return nil, err
				}
				if createdAt.Before(*query.CreatedSince) {
					foundOlderReview = true
					continue
				}
			}

			reviews = append(reviews, review)
			remaining--
			if remaining == 0 {
				break
			}
		}
		included = append(included, resp.Included...)

		if remaining == 0 || foundOlderReview || resp.Links.Next == "" {
			break
		}
		path = resp.Links.Next
	}

	return &CustomerReviewsResult{
		Reviews:             reviews,
		ResponsesByReviewID: mapResponsesByReviewID(reviews, included),
	}, nil
}

func (c *Client) CreateOrUpdateCustomerReviewResponse(reviewID, responseBody string) (*CustomerReviewResponseV1, error) {
	req := CreateCustomerReviewResponseRequest{
		Data: CreateCustomerReviewResponseData{
			Type: "customerReviewResponses",
			Attributes: CreateCustomerReviewResponseAttributes{
				ResponseBody: responseBody,
			},
			Relationships: CreateCustomerReviewResponseRelationships{
				Review: RelationshipData{
					Data: ResourceIdentifier{
						Type: "customerReviews",
						ID:   reviewID,
					},
				},
			},
		},
	}

	data, err := c.do("POST", "/customerReviewResponses", req)
	if err != nil {
		return nil, err
	}

	var resp SingleResponse[CustomerReviewResponseV1]
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &resp.Data, nil
}

func customerReviewsQueryValues(query CustomerReviewsQuery, limit int) url.Values {
	values := url.Values{}
	values.Set("fields[customerReviews]", "rating,title,body,reviewerNickname,createdDate,territory,response")
	values.Set("sort", query.Sort)
	values.Set("limit", strconv.Itoa(limit))

	if len(query.Territories) > 0 {
		values.Set("filter[territory]", strings.Join(query.Territories, ","))
	}
	if len(query.Ratings) > 0 {
		ratings := make([]string, 0, len(query.Ratings))
		for _, rating := range query.Ratings {
			ratings = append(ratings, strconv.Itoa(rating))
		}
		values.Set("filter[rating]", strings.Join(ratings, ","))
	}
	if query.PublishedResponse != nil {
		values.Set("exists[publishedResponse]", strconv.FormatBool(*query.PublishedResponse))
	}
	if query.IncludeResponse {
		values.Set("include", "response")
		values.Set("fields[customerReviewResponses]", "responseBody,lastModifiedDate,state,review")
	}

	return values
}

func customerReviewCreatedAt(review CustomerReview) (time.Time, error) {
	if review.Attributes.CreatedDate == "" {
		return time.Time{}, fmt.Errorf("customer review %s is missing createdDate", review.ID)
	}

	createdAt, err := time.Parse(time.RFC3339, review.Attributes.CreatedDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing createdDate for customer review %s: %w", review.ID, err)
	}
	return createdAt, nil
}

func mapResponsesByReviewID(reviews []CustomerReview, included []CustomerReviewResponseV1) map[string]*CustomerReviewResponseV1 {
	byReviewID := make(map[string]*CustomerReviewResponseV1)
	byResponseID := make(map[string]*CustomerReviewResponseV1)

	for i := range included {
		response := &included[i]
		byResponseID[response.ID] = response

		if response.Relationships == nil {
			continue
		}

		reviewID := response.Relationships.Review.Data.ID
		if reviewID != "" {
			byReviewID[reviewID] = response
		}
	}

	for _, review := range reviews {
		if _, exists := byReviewID[review.ID]; exists {
			continue
		}
		if review.Relationships.Response == nil || review.Relationships.Response.Data == nil {
			continue
		}
		responseID := review.Relationships.Response.Data.ID
		if responseID == "" {
			continue
		}
		if response, exists := byResponseID[responseID]; exists {
			byReviewID[review.ID] = response
		}
	}

	return byReviewID
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
