package models

// TranscribeResponse is the response from the Transcribe API

type TranscribeResponse struct {
	Task     string    `json:"task"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
	Segments []Segment `json:"segments"`
	Text     string    `json:"text"`
}
