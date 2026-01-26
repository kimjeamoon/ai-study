package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"

	"excel-agent/internal/config"
	"excel-agent/internal/processor"

	"github.com/firebase/genkit/go/genkit"
)

type CLI struct {
	Cmd  string
	ID   string
	File string
}

func ParseFlags() *CLI {
	cmd := flag.String("cmd", "", "Command to run: xlsx, sheets, gen")
	id := flag.String("id", "", "Google Spreadsheet ID (for sheets command)")
	file := flag.String("file", "", "JSON file name (for gen command)")
	flag.Parse()

	return &CLI{
		Cmd:  *cmd,
		ID:   *id,
		File: *file,
	}
}

func (c *CLI) Handle(ctx context.Context, g *genkit.Genkit, cfg *config.Config) bool {
	if c.Cmd == "" {
		return false
	}

	switch c.Cmd {
	case "xlsx":
		log.Println("Processing local XLSX files...")
		count, err := processor.ProcessXlsxFiles(cfg.XlsxDir, cfg.JsonDir)
		if err != nil {
			log.Fatalf("XLSX processing failed: %v", err)
		}
		fmt.Printf("Successfully processed %d Excel files.\n", count)

	case "sheets":
		sheetID := c.ID
		if sheetID == "" {
			sheetID = cfg.GoogleSheetID
		}
		if sheetID == "" {
			log.Fatal("Google Spreadsheet ID is required (use -id flag or GOOGLE_SHEET_ID env)")
		}
		log.Printf("Processing Google Sheet ID: %s", sheetID)
		if err := processor.ConvertGoogleSheetToJSON(ctx, sheetID, cfg.JsonDir, cfg.GoogleAPIKey); err != nil {
			log.Fatalf("Google Sheet processing failed: %v", err)
		}
		fmt.Printf("Successfully processed Google Sheet: %s\n", sheetID)

	case "gen":
		log.Printf("Generating Go structs from %s...", c.File)
		res, err := processor.GenerateStructs(ctx, g, c.File, cfg.JsonDir, cfg.DataDir)
		if err != nil {
			log.Fatalf("Struct generation failed: %v", err)
		}
		fmt.Println(res)

	default:
		log.Fatalf("Unknown command: %s. Use xlsx, sheets, or gen.", c.Cmd)
	}

	return true
}
