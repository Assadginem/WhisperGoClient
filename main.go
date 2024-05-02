package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	tts "go-whisper-api/api/texttospeech"
	"go-whisper-api/api/whisper"
	"go-whisper-api/transcribe"
	"go-whisper-api/utils"
	"go-whisper-api/utils/config"
)

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
