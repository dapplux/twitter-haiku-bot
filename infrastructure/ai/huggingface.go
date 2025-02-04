package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dapplux/twitter-haiku-bot/infrastructure/transport"
)

// Hugging Face API rate limits and endpoints.
const (
	huggingFaceMaxRequestsPerMinute = 10 // Free-tier limit
	summaryAPI                      = "https://api-inference.huggingface.co/models/google/pegasus-xsum"
	haikuAPI                        = "https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.2"
)

// HuggingFaceProvider handles AI interactions via Hugging Face API.
type HuggingFaceProvider struct {
	AuthToken string
	Client    *http.Client
}

// NewHuggingFaceProvider initializes a Hugging Face AI provider with rate-limited HTTP transport.
func NewHuggingFaceProvider(authToken string) *HuggingFaceProvider {
	rateLimitedTransport := transport.NewRateLimitTransport(
		float64(huggingFaceMaxRequestsPerMinute)/60, // 10 requests per 60 seconds
		http.DefaultTransport,
	)

	return &HuggingFaceProvider{
		AuthToken: authToken,
		Client:    &http.Client{Transport: rateLimitedTransport},
	}
}

// GenerateSummary uses Pegasus-XSum to summarize text.
func (hf *HuggingFaceProvider) GenerateSummary(text string) (string, error) {
	payload := map[string]string{"inputs": text}
	summary, err := hf.callHuggingFaceModel(summaryAPI, payload)
	if err != nil {
		return "", fmt.Errorf("error in summarization: %v", err)
	}
	return strings.TrimSpace(summary), nil
}

// GenerateHaiku converts a summary into a haiku using Mistral-Small-24B-Instruct-2501.
func (hf *HuggingFaceProvider) GenerateHaiku(summary string) (string, error) {
	prompt := fmt.Sprintf(`Generate a haiku in a strict 5-7-5 syllable format based on the following summary:
	"%s"
	Return only the haiku and nothing else.<RequestEnd>`, summary)

	payload := map[string]string{"inputs": prompt}
	haiku, err := hf.callHuggingFaceModel(haikuAPI, payload)
	if err != nil {
		return "", fmt.Errorf("error in haiku generation: %v", err)
	}
	return extractHaiku(haiku), nil
}

// callHuggingFaceModel makes a POST request to the Hugging Face API with retry logic.
func (hf *HuggingFaceProvider) callHuggingFaceModel(apiURL string, payload map[string]string) (string, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	maxRetries := 5
	backoff := 1 * time.Second
	var lastErr error

	for i := 0; i <= maxRetries; i++ {
		req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return "", err
		}

		req.Header.Set("Authorization", "Bearer "+hf.AuthToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := hf.Client.Do(req)
		if err != nil {
			lastErr = err
		} else {
			bodyBytes, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = readErr
			} else {
				fmt.Println("Raw API Response:", string(bodyBytes))
				// Handle transient errors
				if resp.StatusCode == http.StatusTooManyRequests {
					return "", fmt.Errorf("rate limit exceeded, try again later")
				}
				if resp.StatusCode == http.StatusServiceUnavailable {
					lastErr = fmt.Errorf("service unavailable, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
				} else if resp.StatusCode != http.StatusOK {
					lastErr = fmt.Errorf("request failed, status: %d, body: %s", resp.StatusCode, string(bodyBytes))
				} else {
					// Parse JSON response.
					var arrayResponse []map[string]interface{}
					if err := json.Unmarshal(bodyBytes, &arrayResponse); err == nil && len(arrayResponse) > 0 {
						// Try both possible keys.
						if generatedText, ok := arrayResponse[0]["summary_text"].(string); ok {
							return generatedText, nil
						}
						if generatedText, ok := arrayResponse[0]["generated_text"].(string); ok {
							return generatedText, nil
						}
						lastErr = fmt.Errorf("unexpected response format: %s", string(bodyBytes))
					} else {
						lastErr = fmt.Errorf("could not parse response: %s", string(bodyBytes))
					}
				}
			}
		}

		if i < maxRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	return "", lastErr
}

// extractHaiku removes unnecessary text and extracts the haiku from the generated text.
// It does so by splitting the text into lines and returning the last three non-empty lines.
func extractHaiku(response string) string {
	lines := strings.Split(response, "<RequestEnd>")
	if len(lines) > 1 {
		return lines[1]
	}

	return ""
}
