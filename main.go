package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go-whisper-api/api/whisper"
	"go-whisper-api/transcribe"
	"go-whisper-api/utils"
	"go-whisper-api/utils/config"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// Split text into chunks of up to maxChars
func splitText(text string, maxChars int) []string {
	var parts []string
	runes := []rune(text) // Use runes to avoid splitting in the middle of a character
	for len(runes) > maxChars {
		splitAt := maxChars
		for splitAt > 0 && runes[splitAt-1] != ' ' && runes[splitAt-1] != '.' {
			splitAt-- // Try to avoid splitting in the middle of a word
		}
		if splitAt == 0 {
			splitAt = maxChars // If no spaces, just split at maxChars
		}
		parts = append(parts, string(runes[:splitAt]))
		runes = runes[splitAt:]
	}
	if len(runes) > 0 {
		parts = append(parts, string(runes))
	}
	return parts
}

func OpenAIAPICall(text, apiKey string) ([]byte, error) {
	url := "https://api.openai.com/v1/audio/speech"
	payload := map[string]interface{}{
		"model": "tts-1-hd", // This model name needs to be valid; adjust as necessary
		"input": text,
		"voice": "nova", // Valid voice option
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to generate speech, status: %s, response: %s", resp.Status, string(respBody))
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return audio, nil
}

func performTTS(text, apiKey string) {
	parts := splitText(text, 4096) // Split text into parts of up to 4096 characters
	for index, part := range parts {
		fmt.Printf("Processing part: %s\n", part)
		audio, err := OpenAIAPICall(part, apiKey)
		if err != nil {
			log.Fatalf("Failed to convert text to speech: %v", err)
		}

		// Save the audio to a file
		fileName := fmt.Sprintf("output_audio_part_%d.mp3", index)
		err = os.WriteFile(fileName, audio, 0644)
		if err != nil {
			log.Fatalf("Failed to write audio file %s: %v", fileName, err)
		}
		fmt.Printf("Audio part saved as '%s'.\n", fileName)
	}
	fmt.Println("All parts processed.")
}

func readTextFile(filePath string) (string, error) {
	if !strings.HasSuffix(filePath, ".txt") {
		return "", fmt.Errorf("file format not supported, expected a .txt file")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func main() {
	// Parse the command line arguments
	args, err := utils.ParseArgs()
	if err != nil {
		log.Fatalf("Error parsing arguments: %v", err)
	}

	readConfig, err := config.ReadConfig(args.ConfigPath)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	// Create a new client with the API key
	client := whisper.NewClient(
		whisper.WithKey(readConfig.APIKey),
	)

	switch args.Operation {
	case "transcribe":
		// Perform transcription only
		transcribeConfig := &transcribe.TranscribeConfig{}
		modelOption := transcribe.WithModel(readConfig.Model)
		langOption := transcribe.WithLanguage(args.Language)
		modelOption(transcribeConfig)
		langOption(transcribeConfig)

		response, err := client.TranscribeFile(args.FilePath, modelOption, langOption)
		if err != nil {
			log.Fatalf("Error transcribing file: %v", err)
		}
		fmt.Printf("Transcription: %s\n", response.Text)

	case "tts":
		// Read text directly for tts without transcription
		text, err := readTextFile(args.FilePath)
		if err != nil {
			log.Fatalf("Error reading text file: %v", err)
		}
		performTTS(text, readConfig.APIKey)

	case "both":
		// Transcribe and then convert to speech
		transcribeConfig := &transcribe.TranscribeConfig{}
		modelOption := transcribe.WithModel(readConfig.Model)
		langOption := transcribe.WithLanguage(args.Language)
		modelOption(transcribeConfig)
		langOption(transcribeConfig)

		response, err := client.TranscribeFile(args.FilePath, modelOption, langOption)
		if err != nil {
			log.Fatalf("Error transcribing file: %v", err)
		}
		fmt.Printf("Transcription: %s\n", response.Text)

		performTTS(response.Text, readConfig.APIKey)

	default:
		log.Fatalf("Unsupported operation mode: %s", args.Operation)
	}
}
