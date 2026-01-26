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
	Key  string
}

func ParseFlags() *CLI {
	cmd := flag.String("cmd", "", "Command to run: xlsx, sheets, gen, redis, get")
	id := flag.String("id", "", "Google Spreadsheet ID (for sheets command)")
	file := flag.String("file", "", "JSON file name (for gen command)")
	key := flag.String("key", "", "Redis key name (for get command)")
	flag.Parse()

	return &CLI{
		Cmd:  *cmd,
		ID:   *id,
		File: *file,
		Key:  *key,
	}
}

func (c *CLI) Handle(ctx context.Context, g *genkit.Genkit, cfg *config.Config, reg map[string]interface{}) bool {
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

	case "redis":
		log.Println("Caching JSON data to Redis...")
		if err := processor.CacheJSONToRedis(ctx, cfg.JsonDir, cfg.RedisAddr, cfg.RedisDB); err != nil {
			log.Fatalf("Redis caching failed: %v", err)
		}
		fmt.Println("Successfully cached data to Redis.")

	case "query":
		if c.Key == "" {
			log.Fatal("Query string is required (use -key flag)")
		}
		log.Printf("Querying agent with prompt: %s", c.Key)

		// We use queryFlow which acts as an agent with the redis tool
		if f, ok := reg["queryFlow"].(interface {
			Run(context.Context, string) (string, error)
		}); ok {
			res, err := f.Run(ctx, c.Key)
			if err != nil {
				log.Fatalf("Agent query failed: %v", err)
			}
			fmt.Println(res)
		} else {
			log.Fatal("queryFlow not found in registry")
		}

	default:
		log.Fatalf("Unknown command: %s. Use xlsx, sheets, gen, redis, or get.", c.Cmd)
	}

	return true
}
