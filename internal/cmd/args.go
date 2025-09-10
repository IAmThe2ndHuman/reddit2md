package cmd

import (
	"fmt"
	"os"
)

// Args holds the command-line arguments and Reddit API credentials.
type Args struct {
	Url          string
	OutputDir    string
	BundlePath   string
	Clean        bool
	Delay        int
	Limit        int
	Silent       bool
	ClientId     string
	ClientSecret string
}

// LoadFromEnv loads Reddit API credentials from environment variables.
func (args *Args) LoadFromEnv() {
	args.ClientId = os.Getenv("REDDIT_CLIENT_ID")
	args.ClientSecret = os.Getenv("REDDIT_CLIENT_SECRET")
}

// Validate checks if the provided arguments are valid. (except BundlePath which is checked before bundling)
//
// It also creates the output directory if it does not exist.
func (args *Args) Validate() error {
	if args.ClientId == "" || args.ClientSecret == "" {
		return fmt.Errorf("REDDIT_CLIENT_ID and REDDIT_CLIENT_SECRET environment variables must be set")
	}

	// Create output directory if it doesn't exist
	if args.OutputDir == "" {
		return fmt.Errorf("'output' argument is required")
	}

	if _, err := os.Stat(args.OutputDir); os.IsNotExist(err) {
		err := os.MkdirAll(args.OutputDir, os.ModePerm)

		if err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to access output directory: %v", err)
	}

	if args.Url == "" {
		return fmt.Errorf("'url' argument is required")
	}

	if args.Delay < 0 {
		return fmt.Errorf("delay must be a non-negative integer")
	}

	return nil
}
