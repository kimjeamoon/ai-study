package flows

import (
	"context"
	"fmt"

	"excel-agent/internal/config"
	"excel-agent/internal/processor"

	"github.com/firebase/genkit/go/genkit"
)

func RegisterFlows(g *genkit.Genkit, cfg *config.Config) map[string]interface{} {
	flows := make(map[string]interface{})

	// Define the Flow for local Excel files
	excelFlow := genkit.DefineFlow(g, "excelToJsonFlow", func(ctx context.Context, input string) (string, error) {
		processedCount, err := processor.ProcessXlsxFiles(cfg.XlsxDir, cfg.JsonDir)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Successfully processed %d Excel files.", processedCount), nil
	})
	flows["excelToJsonFlow"] = excelFlow

	// Define the Flow for Google Sheets
	sheetFlow := genkit.DefineFlow(g, "googleSheetToJsonFlow", func(ctx context.Context, spreadsheetID string) (string, error) {
		if spreadsheetID == "" {
			spreadsheetID = cfg.GoogleSheetID
		}

		if spreadsheetID == "" {
			return "", fmt.Errorf("spreadsheetID is required")
		}

		if err := processor.ConvertGoogleSheetToJSON(ctx, spreadsheetID, cfg.JsonDir, cfg.GoogleAPIKey); err != nil {
			return "", err
		}

		return fmt.Sprintf("Successfully processed Google Sheet ID: %s", spreadsheetID), nil
	})
	flows["googleSheetToJsonFlow"] = sheetFlow

	// Define the AI Struct Generator Flow
	structFlow := genkit.DefineFlow(g, "generateStructsFlow", func(ctx context.Context, fileName string) (string, error) {
		return processor.GenerateStructs(ctx, g, fileName, cfg.JsonDir, cfg.DataDir)
	})
	flows["generateStructsFlow"] = structFlow

	return flows
}
