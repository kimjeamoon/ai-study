package flows

import (
	"context"
	"fmt"

	"excel-agent/internal/config"
	"excel-agent/internal/processor"

	"github.com/firebase/genkit/go/genkit"
)

func registerProcessingFlows(g *genkit.Genkit, cfg *config.Config, registry map[string]interface{}) {
	// Local Excel Processor Flow
	registry["excelToJsonFlow"] = genkit.DefineFlow(g, "excelToJsonFlow", func(ctx context.Context, input string) (string, error) {
		processedCount, err := processor.ProcessXlsxFiles(cfg.XlsxDir, cfg.JsonDir)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Successfully processed %d Excel files.", processedCount), nil
	})

	// Google Sheets Processor Flow
	registry["googleSheetToJsonFlow"] = genkit.DefineFlow(g, "googleSheetToJsonFlow", func(ctx context.Context, spreadsheetID string) (string, error) {
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
}
