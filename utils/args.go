package utils

import (
	"flag"
	"fmt"
	"go-whisper-api/models"
	"os"
)

func ParseArgs() (*models.CmdArgs, error) {
	filePath := flag.String("file", "", "The path to the audio file to transcribe")
	language := flag.String("lang", "en", "The language of the audio file")
	configPath := flag.String("config", "config.yaml", "The path to the configuration file")
	operation := flag.String("op", "transcribe", "Operation mode: transcribe, tts, or both")

	flag.Parse()

	if *filePath == "" {
		return nil, fmt.Errorf("The file path is required")
	}
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("the file %s does not exist", *filePath)
	}

	// Validate the operation mode
	validOps := map[string]bool{"transcribe": true, "tts": true, "both": true}
	if _, valid := validOps[*operation]; !valid {
		return nil, fmt.Errorf("invalid operation mode: %s", *operation)
	}

	return &models.CmdArgs{
		FilePath:   *filePath,
		Language:   *language,
		ConfigPath: *configPath,
		Operation:  *operation,
	}, nil
}
