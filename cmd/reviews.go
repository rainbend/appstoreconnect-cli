package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rainbend/appstoreconnect-cli/internal/api"
	"github.com/spf13/cobra"
)

var reviewsCmd = &cobra.Command{
	Use:     "reviews",
	Aliases: []string{"review"},
	Short:   "Query and reply to App Store customer reviews",
}

var reviewsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"query", "ls"},
	Short:   "List App Store customer reviews",
	RunE:    runReviewsList,
}

var reviewsReplyCmd = &cobra.Command{
	Use:     "reply",
	Aliases: []string{"respond"},
	Short:   "Create or update a reply to a customer review",
	RunE:    runReviewsReply,
}

func init() {
	reviewsListCmd.Flags().StringP("app-id", "a", os.Getenv("ASC_APP_ID"), "App Store Connect App ID (env: ASC_APP_ID)")
	reviewsListCmd.Flags().String("territory", "", "Territory filter, comma-separated App Store territory codes such as USA, CHN, HKG")
	reviewsListCmd.Flags().String("rating", "", "Rating filter, comma-separated values from 1 to 5")
	reviewsListCmd.Flags().String("response-state", "all", "Published response filter: all, responded, or unresponded")
	reviewsListCmd.Flags().String("since", "", "Only include reviews created at or after this date/time (YYYY-MM-DD or RFC3339)")
	reviewsListCmd.Flags().Int("days", 0, "Only include reviews from the last N days, such as 7 for the last week")
	reviewsListCmd.Flags().String("sort", "-createdDate", "Sort: createdDate, -createdDate, rating, or -rating")
	reviewsListCmd.Flags().Int("limit", 20, "Maximum number of reviews to return")
	reviewsListCmd.Flags().Bool("include-response", false, "Include response body and state when available")
	reviewsListCmd.Flags().Bool("json", false, "Print JSON output")

	reviewsReplyCmd.Flags().String("review-id", "", "Customer review ID to reply to")
	reviewsReplyCmd.Flags().String("body", "", "Reply text")
	reviewsReplyCmd.Flags().String("file", "", "Path to a file containing reply text, or '-' to read stdin")
	reviewsReplyCmd.Flags().Bool("json", false, "Print JSON output")
	_ = reviewsReplyCmd.MarkFlagRequired("review-id")

	reviewsCmd.AddCommand(reviewsListCmd)
	reviewsCmd.AddCommand(reviewsReplyCmd)
	rootCmd.AddCommand(reviewsCmd)
}

func runReviewsList(cmd *cobra.Command, args []string) error {
	if err := validateAuthFlags(); err != nil {
		return err
	}

	appID, _ := cmd.Flags().GetString("app-id")
	if appID == "" {
		return fmt.Errorf("--app-id or ASC_APP_ID environment variable is required")
	}

	territoryFlag, _ := cmd.Flags().GetString("territory")
	territories := parseCommaSeparatedUpper(territoryFlag)

	ratingFlag, _ := cmd.Flags().GetString("rating")
	ratings, err := parseRatings(ratingFlag)
	if err != nil {
		return err
	}

	responseState, _ := cmd.Flags().GetString("response-state")
	publishedResponse, err := parsePublishedResponseFilter(responseState)
	if err != nil {
		return err
	}

	sort, _ := cmd.Flags().GetString("sort")
	if err := validateReviewSort(sort); err != nil {
		return err
	}

	sinceFlag, _ := cmd.Flags().GetString("since")
	days, _ := cmd.Flags().GetInt("days")
	createdSince, err := parseReviewCreatedSince(sinceFlag, days, time.Now())
	if err != nil {
		return err
	}
	if createdSince != nil && sort != "-createdDate" {
		return fmt.Errorf("--since and --days require --sort -createdDate")
	}

	limit, _ := cmd.Flags().GetInt("limit")
	if limit <= 0 {
		return fmt.Errorf("--limit must be greater than 0")
	}

	includeResponse, _ := cmd.Flags().GetBool("include-response")
	if publishedResponse != nil && *publishedResponse {
		includeResponse = true
	}

	client := api.NewClient(keyID, issuerID, privateKeyPath)
	result, err := client.ListCustomerReviews(appID, api.CustomerReviewsQuery{
		Territories:       territories,
		Ratings:           ratings,
		PublishedResponse: publishedResponse,
		IncludeResponse:   includeResponse,
		Sort:              sort,
		Limit:             limit,
		CreatedSince:      createdSince,
	})
	if err != nil {
		return fmt.Errorf("listing customer reviews: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		return printReviewsJSON(result)
	}

	if len(result.Reviews) == 0 {
		fmt.Println("No reviews found.")
		return nil
	}

	fmt.Printf("Found %d review(s).\n\n", len(result.Reviews))
	for _, review := range result.Reviews {
		printReview(review, result.ResponsesByReviewID[review.ID])
	}

	return nil
}

func runReviewsReply(cmd *cobra.Command, args []string) error {
	if err := validateAuthFlags(); err != nil {
		return err
	}

	reviewID, _ := cmd.Flags().GetString("review-id")
	body, _ := cmd.Flags().GetString("body")
	filePath, _ := cmd.Flags().GetString("file")
	if body != "" && filePath != "" {
		return fmt.Errorf("use either --body or --file, not both")
	}
	if filePath != "" {
		data, err := readReplyFile(filePath)
		if err != nil {
			return err
		}
		body = string(data)
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("--body or --file is required and must not be empty")
	}

	client := api.NewClient(keyID, issuerID, privateKeyPath)
	response, err := client.CreateOrUpdateCustomerReviewResponse(reviewID, body)
	if err != nil {
		return fmt.Errorf("creating or updating customer review response: %w", err)
	}

	jsonOutput, _ := cmd.Flags().GetBool("json")
	if jsonOutput {
		return printJSON(response)
	}

	fmt.Printf("Response updated for review %s\n", reviewID)
	if response.ID != "" {
		fmt.Printf("Response ID: %s\n", response.ID)
	}
	if response.Attributes.State != "" {
		fmt.Printf("State: %s\n", response.Attributes.State)
	}
	if response.Attributes.LastModifiedDate != "" {
		fmt.Printf("Modified: %s\n", response.Attributes.LastModifiedDate)
	}

	return nil
}

func readReplyFile(path string) ([]byte, error) {
	if path == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading reply from stdin: %w", err)
		}
		return data, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading reply file: %w", err)
	}
	return data, nil
}

func parseCommaSeparatedUpper(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToUpper(strings.TrimSpace(part))
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func parseRatings(value string) ([]int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}

	parts := strings.Split(value, ",")
	ratings := make([]int, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		rating, err := strconv.Atoi(part)
		if err != nil || rating < 1 || rating > 5 {
			return nil, fmt.Errorf("invalid --rating %q: values must be numbers from 1 to 5", value)
		}
		ratings = append(ratings, rating)
	}
	return ratings, nil
}

func parsePublishedResponseFilter(value string) (*bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "all":
		return nil, nil
	case "responded", "replied", "published", "yes", "true":
		v := true
		return &v, nil
	case "unresponded", "unreplied", "unpublished", "no", "false":
		v := false
		return &v, nil
	default:
		return nil, fmt.Errorf("invalid --response-state %q: must be all, responded, or unresponded", value)
	}
}

func validateReviewSort(value string) error {
	switch value {
	case "createdDate", "-createdDate", "rating", "-rating":
		return nil
	default:
		return fmt.Errorf("invalid --sort %q: must be createdDate, -createdDate, rating, or -rating", value)
	}
}

func parseReviewCreatedSince(since string, days int, now time.Time) (*time.Time, error) {
	since = strings.TrimSpace(since)
	if days < 0 {
		return nil, fmt.Errorf("--days must not be negative")
	}
	if since != "" && days > 0 {
		return nil, fmt.Errorf("use either --since or --days, not both")
	}
	if days > 0 {
		value := now.AddDate(0, 0, -days)
		return &value, nil
	}
	if since == "" {
		return nil, nil
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		value, err := time.Parse(layout, since)
		if err == nil {
			return &value, nil
		}
	}

	value, err := time.ParseInLocation("2006-01-02", since, time.Local)
	if err != nil {
		return nil, fmt.Errorf("invalid --since %q: use YYYY-MM-DD or RFC3339", since)
	}
	return &value, nil
}

func printReview(review api.CustomerReview, response *api.CustomerReviewResponseV1) {
	attrs := review.Attributes
	fmt.Printf("Review ID: %s\n", review.ID)
	if attrs.CreatedDate != "" {
		fmt.Printf("Created: %s\n", attrs.CreatedDate)
	}
	if attrs.Territory != "" {
		fmt.Printf("Territory: %s\n", attrs.Territory)
	}
	if attrs.ReviewerNickname != "" {
		fmt.Printf("Reviewer: %s\n", attrs.ReviewerNickname)
	}
	fmt.Printf("Rating: %d\n", attrs.Rating)
	if attrs.Title != "" {
		printMultiline("Title", attrs.Title)
	}
	if attrs.Body != "" {
		printMultiline("Body", attrs.Body)
	}
	if response != nil {
		fmt.Printf("Response ID: %s\n", response.ID)
		if response.Attributes.State != "" {
			fmt.Printf("Response State: %s\n", response.Attributes.State)
		}
		if response.Attributes.LastModifiedDate != "" {
			fmt.Printf("Response Modified: %s\n", response.Attributes.LastModifiedDate)
		}
		if response.Attributes.ResponseBody != "" {
			printMultiline("Response Body", response.Attributes.ResponseBody)
		}
	}
	fmt.Println()
}

func printMultiline(label, value string) {
	lines := strings.Split(value, "\n")
	fmt.Printf("%s: %s\n", label, lines[0])
	for _, line := range lines[1:] {
		fmt.Printf("  %s\n", line)
	}
}

type reviewOutput struct {
	ID               string          `json:"id"`
	Rating           int             `json:"rating"`
	Title            string          `json:"title,omitempty"`
	Body             string          `json:"body,omitempty"`
	ReviewerNickname string          `json:"reviewerNickname,omitempty"`
	CreatedDate      string          `json:"createdDate,omitempty"`
	Territory        string          `json:"territory,omitempty"`
	Response         *responseOutput `json:"response,omitempty"`
}

type responseOutput struct {
	ID               string `json:"id"`
	ResponseBody     string `json:"responseBody,omitempty"`
	LastModifiedDate string `json:"lastModifiedDate,omitempty"`
	State            string `json:"state,omitempty"`
}

func printReviewsJSON(result *api.CustomerReviewsResult) error {
	output := make([]reviewOutput, 0, len(result.Reviews))
	for _, review := range result.Reviews {
		attrs := review.Attributes
		item := reviewOutput{
			ID:               review.ID,
			Rating:           attrs.Rating,
			Title:            attrs.Title,
			Body:             attrs.Body,
			ReviewerNickname: attrs.ReviewerNickname,
			CreatedDate:      attrs.CreatedDate,
			Territory:        attrs.Territory,
		}
		if response := result.ResponsesByReviewID[review.ID]; response != nil {
			item.Response = &responseOutput{
				ID:               response.ID,
				ResponseBody:     response.Attributes.ResponseBody,
				LastModifiedDate: response.Attributes.LastModifiedDate,
				State:            response.Attributes.State,
			}
		}
		output = append(output, item)
	}
	return printJSON(output)
}

func printJSON(value interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
