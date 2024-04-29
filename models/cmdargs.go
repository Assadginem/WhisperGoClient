package models

// CmdArgs holds the command-line arguments provided by the user.
type CmdArgs struct {
	FilePath   string
	Language   string
	ConfigPath string
	Operation  string
}
