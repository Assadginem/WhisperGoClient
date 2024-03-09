package transcribe

type TranscribeConfig struct {
	Model    string
	Language string
	File     string
}

type TranscribeOption func(*TranscribeConfig)

func WithModel(model string) TranscribeOption {
	return func(c *TranscribeConfig) {
		c.Model = model
	}
}

func WithLanguage(language string) TranscribeOption {
	return func(c *TranscribeConfig) {
		c.Language = language
	}
}

func WithFile(file string) TranscribeOption {
	return func(c *TranscribeConfig) {
		c.File = file
	}
}
