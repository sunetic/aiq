package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aiq/aiq/internal/cli"
	"github.com/aiq/aiq/internal/config"
	"github.com/aiq/aiq/internal/prompt"
	"github.com/aiq/aiq/internal/source"
	"github.com/aiq/aiq/internal/sql"
	"github.com/aiq/aiq/internal/ui"
	"github.com/aiq/aiq/internal/version"
)

func main() {
	// Handle version command flags before any initialization
	for _, arg := range os.Args[1:] {
		if arg == "-v" || arg == "--version" {
			fmt.Println(version.GetVersionInfo())
			os.Exit(0)
		}
	}

	// Ensure directory structure exists (needed for prompt initialization)
	if err := config.EnsureDirectoryStructure(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create config directory structure: %v\n", err)
		os.Exit(1)
	}

	// Initialize prompt loader to ensure default prompt files are created
	// This happens at startup so prompts are available even if user doesn't enter chat mode
	_, err := prompt.NewLoader()
	if err != nil {
		// Log warning but don't fail - prompts will use fallback defaults
		fmt.Printf("Warning: Failed to initialize prompts: %v. Using default prompts.\n", err)
	}

	// Parse database connection arguments first (before flag.Parse to avoid conflicts)
	dbArgs, err := cli.ParseDatabaseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Parse session flag separately (only if no database args)
	var sessionFile string
	if dbArgs == nil {
		flag.StringVar(&sessionFile, "s", "", "Path to session file to restore")
		flag.StringVar(&sessionFile, "session", "", "Path to session file to restore")
		flag.Parse()
	} else {
		// Parse session flag manually from args (for compatibility with database args)
		for i, arg := range os.Args[1:] {
			if (arg == "-s" || arg == "--session") && i+1 < len(os.Args[1:]) {
				sessionFile = os.Args[i+2]
				break
			}
			if strings.HasPrefix(arg, "-s") && len(arg) > 2 {
				sessionFile = arg[2:]
				break
			}
			if strings.HasPrefix(arg, "--session=") {
				sessionFile = strings.TrimPrefix(arg, "--session=")
				break
			}
		}
	}

	// If database args provided, handle direct connection
	if dbArgs != nil {
		// Validate connection
		if err := cli.ValidateConnection(dbArgs); err != nil {
			fmt.Fprintf(os.Stderr, "Connection failed: %v\n", err)
			os.Exit(1)
		}

		// Create source with auto-generated name
		newSource := &source.Source{
			Type:     dbArgs.Engine,
			Host:     dbArgs.Host,
			Port:     dbArgs.Port,
			Database: dbArgs.Database,
			Username: dbArgs.Username,
			Password: dbArgs.Password,
		}

		// Check if source already exists before creating
		existingName, err := source.FindExistingSourceByConnection(dbArgs.Host, dbArgs.Port, dbArgs.Username)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check existing sources: %v\n", err)
			os.Exit(1)
		}

		var sourceName string
		if existingName != "" {
			// Source already exists, use it
			sourceName = existingName
			ui.ShowInfo(fmt.Sprintf("Connected to database. Using existing source '%s'.", sourceName))
		} else {
			// Create new source
			sourceName, err = source.AddSourceWithAutoName(newSource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create source: %v\n", err)
				os.Exit(1)
			}
			ui.ShowInfo(fmt.Sprintf("Connected to database. Source '%s' created.", sourceName))
		}

		// Directly enter chat mode with the created source
		// Pass Database from dbArgs as overrideDatabase to use for this session only
		if err := sql.RunSQLModeWithSource(sourceName, sessionFile, dbArgs.Database); err != nil {
			// Check if error is ErrReturnToMenu - this is expected when user exits chat mode
			if err == sql.ErrReturnToMenu {
				// Normal return, exit gracefully
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// No database args, use normal flow
	if err := cli.Run(sessionFile); err != nil {
		os.Exit(1)
	}
}
