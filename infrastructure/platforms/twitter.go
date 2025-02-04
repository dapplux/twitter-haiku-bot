package platforms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/dapplux/twitter-haiku-bot/entities"
	"github.com/dapplux/twitter-haiku-bot/infrastructure/transport"
	"github.com/dghubble/oauth1"
)

// Twitter API Constants (Free API Rate Limits)
const (
	TwitterBaseURL        = "https://api.twitter.com/2"
	TwitterSearchEndpoint = TwitterBaseURL + "/tweets/search/recent"
	TwitterPostEndpoint   = TwitterBaseURL + "/tweets"
	TwitterFreeAPILimit   = 10               // Free API allows 10 requests per 15 minutes
	TwitterRateLimitReset = 15 * time.Minute // API resets every 15 minutes
)

// TwitterProvider manages API interactions using an HTTP client.
type TwitterProvider struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
	Client            *http.Client
}

// NewTwitterProvider initializes a Twitter API client with OAuth 1.0a and rate limiting.
func NewTwitterProvider(consumerKey, consumerSecret, accessToken, accessTokenSecret string) *TwitterProvider {
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	// Create the OAuth1-signed client.
	httpClient := config.Client(oauth1.NoContext, token)
	// Wrap the existing OAuth client's transport with a rate-limited transport.
	httpClient.Transport = transport.NewRateLimitTransport(
		float64(TwitterFreeAPILimit)/TwitterRateLimitReset.Seconds(), // Rate: 10 requests per 900 sec
		httpClient.Transport,
	)

	return &TwitterProvider{
		ConsumerKey:       consumerKey,
		ConsumerSecret:    consumerSecret,
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
		Client:            httpClient,
	}
}

// FetchPosts fetches recent tweets matching a query using OAuth1 authentication.
func (tp *TwitterProvider) FetchPosts(ctx context.Context, limit int) ([]entities.Post, error) {
	// Build query parameters.
	q := url.Values{}
	q.Set("query", "software news lang:en -is:retweet")
	q.Set("max_results", fmt.Sprintf("%d", limit))
	q.Set("sort_order", "relevancy") // Sort by popularity.
	q.Set("expansions", "author_id") // Expand author ID.
	q.Set("tweet.fields", "public_metrics,author_id")
	q.Set("user.fields", "username") // Get usernames in `includes.users`.

	apiURL := fmt.Sprintf("%s?%s", TwitterSearchEndpoint, q.Encode())

	maxRetries := 3
	backoff := 1 * time.Second
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		// Remove the Bearer token header since OAuth1-signed client adds its own Authorization header.
		req.Header.Set("Content-Type", "application/json")

		resp, err := tp.Client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				lastErr = err
			} else {
				// Check for rate limiting.
				if resp.StatusCode == http.StatusTooManyRequests {
					resetHeader := resp.Header.Get("x-rate-limit-reset")
					if resetHeader != "" {
						if resetTimestamp, err := strconv.ParseInt(resetHeader, 10, 64); err == nil {
							waitDuration := time.Until(time.Unix(resetTimestamp, 0))
							if waitDuration > 0 {
								fmt.Printf("Rate limit hit. Waiting for %v until reset.\n", waitDuration)
								time.Sleep(waitDuration)
							}
						} else {
							fmt.Printf("Error parsing x-rate-limit-reset header: %v. Falling back to %v backoff.\n", err, backoff)
							time.Sleep(backoff)
							backoff *= 2
						}
					} else {
						fmt.Printf("x-rate-limit-reset header not found. Retrying in %v.\n", backoff)
						time.Sleep(backoff)
						backoff *= 2
					}
					lastErr = fmt.Errorf("too many requests: %s", string(bodyBytes))
				} else if resp.StatusCode != http.StatusOK {
					lastErr = fmt.Errorf("failed to fetch tweets: %s", string(bodyBytes))
				} else {
					var result map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &result); err != nil {
						return nil, err
					}
					return mapTwitterPosts(result)
				}
			}
		}

		if i < maxRetries && lastErr != nil {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return nil, fmt.Errorf("error fetching posts from platform: %v", lastErr)
}

// CommentOn replies to a tweet with a given message using OAuth 1.0a.
func (tp *TwitterProvider) CommentOn(tweetID, message string) error {
	payload := map[string]interface{}{
		"text": message,
		"reply": map[string]string{
			"in_reply_to_tweet_id": tweetID,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", TwitterPostEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	// The OAuth1-signed client automatically adds the required Authorization header.

	resp, err := tp.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println("Raw Response:", string(bodyBytes))
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to comment on tweet: %s", string(bodyBytes))
	}

	return nil
}

// mapTwitterPosts maps the JSON response to a slice of entities.Post.
// If there are no tweets, it returns an empty slice.
func mapTwitterPosts(result map[string]interface{}) ([]entities.Post, error) {
	rawTweets, ok := result["data"].([]interface{})
	if !ok {
		// If there's no "data" key, assume no tweets were found.
		return []entities.Post{}, nil
	}

	// Build a map from user ID to username.
	userMap := make(map[string]string)
	if includes, exists := result["includes"].(map[string]interface{}); exists {
		if users, ok := includes["users"].([]interface{}); ok {
			for _, rawUser := range users {
				user, _ := rawUser.(map[string]interface{})
				id, _ := user["id"].(string)
				username, _ := user["username"].(string)
				userMap[id] = username
			}
		}
	}

	var posts []entities.Post
	for _, rawTweet := range rawTweets {
		tweet, ok := rawTweet.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := tweet["id"].(string)
		authorID, _ := tweet["author_id"].(string)
		text, _ := tweet["text"].(string)
		createdAtStr, _ := tweet["created_at"].(string)

		var createdAt time.Time
		if createdAtStr != "" {
			parsedTime, err := time.Parse(time.RFC3339, createdAtStr)
			if err == nil {
				createdAt = parsedTime
			}
		}

		metrics, _ := tweet["public_metrics"].(map[string]interface{})
		likes := intOrDefault(metrics, "like_count")
		shares := intOrDefault(metrics, "retweet_count")
		replies := intOrDefault(metrics, "reply_count")
		username := userMap[authorID]

		posts = append(posts, entities.Post{
			ID: id,
			Author: entities.Author{
				ID:       authorID,
				Username: username,
			},
			Text:      text,
			Likes:     likes,
			Shares:    shares,
			Replies:   replies,
			Platform:  entities.PlatformTwitter,
			CreatedAt: createdAt,
		})
	}

	return posts, nil
}

func intOrDefault(metrics map[string]interface{}, key string) int {
	if value, ok := metrics[key].(float64); ok {
		return int(value)
	}
	return 0
}
