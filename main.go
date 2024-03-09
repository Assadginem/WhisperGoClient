package main

import (
	"fmt"
	"go-whisper-api/api/whisper"
	"go-whisper-api/transcribe"
	"go-whisper-api/utils"
	"go-whisper-api/utils/config"
	"log"
)

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

	// Create a new client with the API key and the language
	client := whisper.NewClient(
		whisper.WithKey(readConfig.APIKey),
	)

	// Create a new TranscribeConfig with the model
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
}
