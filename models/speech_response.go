package models

// SpeechResponse represents the response structure for the TTS API.
type SpeechResponse struct {
	AudioContent []byte  // This will hold the raw audio data.
	Duration     float64 `json:"duration"` // Duration of the audio clip in seconds.
	Format       string  `json:"format"`   // Format of the audio file, e.g., 'mp3'.
}
