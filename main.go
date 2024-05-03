package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	tts "go-whisper-api/api/texttospeech"
	"go-whisper-api/api/whisper"
	"go-whisper-api/transcribe"
	"go-whisper-api/utils"
	"go-whisper-api/utils/config"
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
	// Initialize necessary services and configurations
	args, err := utils.ParseArgs()
	if err != nil {
		log.Fatalf("Error parsing arguments: %v", err)
	}

	readConfig, err := config.ReadConfig(args.ConfigPath)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	whisperClient := whisper.NewClient(
		whisper.WithKey(readConfig.APIKey),
	)

	// Prepare and execute transcription
	transcribeConfig := &transcribe.TranscribeConfig{}
	modelOption := transcribe.WithModel(readConfig.Model)
	langOption := transcribe.WithLanguage(args.Language)
	modelOption(transcribeConfig)
	langOption(transcribeConfig)

	response, err := whisperClient.TranscribeFile(args.FilePath, modelOption, langOption)
	if err != nil {
		log.Fatalf("Error transcribing file: %v", err)
	}
	fmt.Printf("Transcription: %s\n", response.Text)

	// Set up and use the TTS client with the transcription result
	ttsClient := tts.NewClient(
		tts.WithKey(readConfig.APIKey),
	)

	ttsOptions := tts.SpeakOptions{
		Model:          "tts-1",
		Voice:          "nova",
		ResponseFormat: "mp3",
	}

	// Print the transcription text and TTS options for debugging
	fmt.Println("Transcription Text:", response.Text)
	fmt.Println("TTS Options:", ttsOptions)

	speechResponses, err := ttsClient.Speak(response.Text, ttsOptions)
	if err != nil {
		log.Fatalf("Error converting text to speech: %v", err)
	}

	// Check if the TTS API response is empty
	if len(speechResponses) == 0 {
		log.Fatal("Empty speech responses from TTS API")
	}

	// Concatenate audio content from all responses
	var buffer bytes.Buffer
	for _, resp := range speechResponses {
		if _, err := buffer.Write(resp.AudioContent); err != nil {
			log.Fatalf("Error writing audio content to buffer: %v", err)
		}
	}

	// Write concatenated audio content to a file
	outputFileName := "output_audio.mp3"
	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("Error creating output audio file: %v", err)
	}
	defer outputFile.Close()

	if _, err := io.Copy(outputFile, &buffer); err != nil {
		log.Fatalf("Error writing buffer content to output audio file: %v", err)
	}

	fmt.Printf("Concatenated audio saved as '%s'.\n", outputFileName)
}
