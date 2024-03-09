package whisper

import "github.com/jarcoal/httpmock"

import (
	"net/http"
	"testing"
)

// mockHTTP for mocking HTTP responses
func mockHTTP(statusCode int, responseBody string) {
	httpmock.Activate()
	httpmock.RegisterResponder("POST", "https://api.openai.com/v1/audio/transcriptions",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(statusCode, responseBody)
			return resp, nil
		},
	)
}

// TestTranscribeFileSuccess tests the successful transcription of a file
func TestTranscribeFileSuccess(t *testing.T) {
	mockHTTP(200, `{"text": "sample transcribed text"}`)

	client := NewClient(WithKey("test_api_key"))
	response, err := client.TranscribeFile("path/to/test/file")

	httpmock.DeactivateAndReset()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedText := "sample transcribed text"
	if response.Text != expectedText {
		t.Errorf("Expected transcription text to be '%s', got '%s'", expectedText, response.Text)
	}
}

// TestTranscribeFileError tests handling of an error during transcription
func TestTranscribeFileError(t *testing.T) {
	mockHTTP(500, "Internal Server Error")

	client := NewClient(WithKey("test_api_key"))
	_, err := client.TranscribeFile("path/to/test/file")

	httpmock.DeactivateAndReset()

	if err == nil {
		t.Fatalf("Expected an error, got none")
	}
}

// TestTranscribeFileNoAPIKey tests error handling when no API key is provided
func TestTranscribeFileNoAPIKey(t *testing.T) {
	// Assuming NewClient returns an error or nil client when no API key is provided
	client := NewClient() // No API key provided
	if client != nil {
		_, err := client.TranscribeFile("path/to/test/file")
		if err == nil {
			t.Fatalf("Expected an error due to missing API key, got none")
		}
	}
}
