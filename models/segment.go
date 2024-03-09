package models

type Segment struct {
	ID               int     `json:"id"`
	Seek             int     `json:"seek"`
	Start            float64 `json:"start"`
	End              float64 `json:"end"`
	Text             string  `json:"text"`
	Token            []int   `json:"token"`
	Temperature      float64 `json:"temperature"`
	AvgLogProb       float64 `json:"avg_logprob"`
	CompressionRatio float64 `json:"compressionratio"`
	NoSpeechProb     float64 `json:"no_speech_prob"`
	Transient        bool    `json:"transient"`
}
