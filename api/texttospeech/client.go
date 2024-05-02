package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-whisper-api/models"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	DefaultBaseURL = "https://api.openai.com/v1"
	MaxTextLength  = 4096 // Maximum characters allowed in the input text
)

// Client is the main structure for interacting with the OpenAI Text-to-Speech API.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// ClientOption is a function type that allows setting options for the Client.
type ClientOption func(*Client)

// WithKey sets the API key for the Client.
func WithKey(key string) ClientOption {
	return func(c *Client) {
		c.apiKey = key
	}
}

// WithBaseURL sets the base URL for the Client.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets the HTTP client for the Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// NewClient creates a new TTS API client with the given options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL: DefaultBaseURL, // Default base URL
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.apiKey == "" {
		c.apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}

	return c
}

// SpeakOptions struct for passing options to the Speak method.
type SpeakOptions struct {
	Model          string
	Voice          string
	ResponseFormat string
}

// SplitText splits the input text into chunks that do not exceed the maximum length.
func SplitText(text string) []string {
	var parts []string
	words := strings.Fields(text)
	var currentPart strings.Builder

	for _, word := range words {
		if currentPart.Len()+len(word)+1 > MaxTextLength {
			parts = append(parts, currentPart.String())
			currentPart.Reset()
		}
		if currentPart.Len() > 0 {
			currentPart.WriteByte(' ')
		}
		currentPart.WriteString(word)
	}
	if currentPart.Len() > 0 {
		parts = append(parts, currentPart.String())
	}
	return parts
}

// Speak converts the given text to speech using the OpenAI TTS API.
func (c *Client) Speak(text string, opts SpeakOptions) ([]*models.SpeechResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API key is missing. Set the OPENAI_API_KEY environment variable")
	}

	parts := SplitText(text)
	var responses []*models.SpeechResponse

	for _, part := range parts {
		payload := map[string]interface{}{
			"model":           opts.Model,
			"input":           part,
			"voice":           opts.Voice,
			"response_format": opts.ResponseFormat,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %v", err)
		}

		req, err := http.NewRequest("POST", c.URL("/audio/speech"), bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %v", err)
		}
		defer resp.Body.Close()

		// Check the HTTP status code
		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to generate speech, status: %s, response: %s", resp.Status, string(respBody))
		}

		// Read audio data
		audioData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read audio data: %v", err)
		}

		// Append audio data to responses
		responses = append(responses, &models.SpeechResponse{AudioContent: audioData})
	}

	return responses, nil
}

// URL constructs the full URL for the given relative path.
func (c *Client) URL(relPath string) string {
	return fmt.Sprintf("%s%s", c.baseURL, relPath)
}
