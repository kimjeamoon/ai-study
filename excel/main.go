package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/joho/godotenv"
	"github.com/xuri/excelize/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	// Init Genkit with Ollama plugin
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/gemini-2.5-flash"),
	)

	// Define the Flow for local Excel files
	genkit.DefineFlow(g, "excelToJsonFlow", func(ctx context.Context, input string) (string, error) {
		// input can be ignored or used to specify a specific file/folder,
		// but per requirements we scan 'xlsx' folder.

		xlsxDir := "xlsx"
		jsonDir := "json"

		// Ensure json directory exists
		if err := os.MkdirAll(jsonDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create json directory: %w", err)
		}

		files, err := os.ReadDir(xlsxDir)
		if err != nil {
			return "", fmt.Errorf("failed to read xlsx directory: %w", err)
		}

		processedCount := 0
		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".xlsx" {
				continue
			}

			filePath := filepath.Join(xlsxDir, file.Name())
			if err := convertExcelToJSON(filePath, jsonDir); err != nil {
				log.Printf("Failed to convert %s: %v", file.Name(), err)
				continue
			}
			processedCount++
		}

		return fmt.Sprintf("Successfully processed %d Excel files.", processedCount), nil
	})

	// Define the Flow for Google Sheets
	googleSheetFlow := genkit.DefineFlow(g, "googleSheetToJsonFlow", func(ctx context.Context, spreadsheetID string) (string, error) {
		if spreadsheetID == "" {
			spreadsheetID = os.Getenv("GOOGLE_SHEET_ID")
		}

		if spreadsheetID == "" {
			return "", fmt.Errorf("spreadsheetID is required")
		}

		jsonDir := "json"
		if err := os.MkdirAll(jsonDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create json directory: %w", err)
		}

		if err := convertGoogleSheetToJSON(ctx, spreadsheetID, jsonDir); err != nil {
			return "", err
		}

		return fmt.Sprintf("Successfully processed Google Sheet ID: %s", spreadsheetID), nil
	})

	// Define the AI Struct Generator Flow
	generateStructsFlow := genkit.DefineFlow(g, "generateStructsFlow", func(ctx context.Context, fileName string) (string, error) {
		jsonDir := "json"
		if fileName == "" {
			// Find the first JSON file in json directory
			files, err := os.ReadDir(jsonDir)
			if err != nil || len(files) == 0 {
				return "", fmt.Errorf("no JSON files found in %s", jsonDir)
			}
			for _, f := range files {
				if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
					fileName = f.Name()
					break
				}
			}
		}

		if fileName == "" {
			return "", fmt.Errorf("could not find a JSON file to process")
		}

		jsonPath := filepath.Join(jsonDir, fileName)
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			return "", fmt.Errorf("failed to read JSON file: %v", err)
		}

		// Parse just enough to get the schema
		var jsonRaw map[string]interface{}
		if err := json.Unmarshal(data, &jsonRaw); err != nil {
			return "", fmt.Errorf("failed to parse JSON: %v", err)
		}

		// Create a sample to send to the AI
		sample := make(map[string]interface{})
		for sheetName, content := range jsonRaw {
			rows, ok := content.([]interface{})
			if ok && len(rows) > 0 {
				sample[sheetName] = rows[0]
			}
		}

		baseName := filepath.Base(fileName)
		sampleData, _ := json.MarshalIndent(sample, "", "  ")

		prompt := fmt.Sprintf(`Generate Go structs based on the following JSON sample. 
The JSON represents a spreadsheet where each top-level key is a sheet name, 
and its value is an array of objects.
Use the keys in the sample objects to define the struct fields.
Exclude and ignore any sheet names or object keys that contain Korean characters (Hangul).
Include standard JSON tags.
Name the individual structs after the sheet names (converted to PascalCase).
Define a container struct named [%s]AllSheets. 
(Replace [%s] with the actual JSON source name in PascalCase).
For each sheet, create a field in this container struct using the sheet name (converted to PascalCase) as the field name, 
and set its type as a slice of the corresponding struct.
Output ONLY the Go code, starting with 'package data'.

JSON Sample:
%s`, baseName, baseName, string(sampleData))

		resp, err := genkit.GenerateText(ctx, g, ai.WithPrompt(prompt))
		if err != nil {
			return "", fmt.Errorf("AI generation failed: %v", err)
		}

		// Clean up markdown block if present
		code := resp
		code = strings.TrimPrefix(code, "```go")
		code = strings.TrimPrefix(code, "```")
		code = strings.TrimSuffix(code, "```")
		code = strings.TrimSpace(code)
		code = strings.TrimSuffix(code, "```")

		goFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".go"
		if err := os.WriteFile("data/"+goFileName, []byte(code), 0644); err != nil {
			return "", fmt.Errorf("failed to write %s: %v", goFileName, err)
		}

		return "Successfully generated " + goFileName + " using AI", nil
	})

	// For verification: run immediately if env var is set
	if os.Getenv("RUN_ON_START") == "true" {
		log.Println("Running flow immediately...")
		// Check if we want to run google sheet flow
		if sheetID := os.Getenv("GOOGLE_SHEET_ID"); sheetID != "" {
			res, err := googleSheetFlow.Run(ctx, sheetID)
			if err != nil {
				log.Fatalf("Google Sheet Flow execution failed: %v", err)
			}
			fmt.Println(res)
			return
		}

		// Check if we want to run AI struct generator flow
		if os.Getenv("GOOGLE_GEN_STRUCTS") == "true" {
			res, err := generateStructsFlow.Run(ctx, "")
			if err != nil {
				log.Fatalf("Generate Structs Flow execution failed: %v", err)
			}
			fmt.Println(res)
			return
		}

		log.Println("RUN_ON_START supported. Set GOOGLE_SHEET_ID or GOOGLE_GEN_STRUCTS=true, or use genkit start.")
	}

	log.Println("Genkit agent started. Waiting for requests...")
	// Block until context is cancelled (via SIGINT/SIGTERM handled by genkit.Init)
	<-ctx.Done()
}

func convertGoogleSheetToJSON(ctx context.Context, spreadsheetID, jsonDir string) error {
	var opts []option.ClientOption

	if _, err := os.Stat("credentials.json"); err == nil {
		opts = append(opts, option.WithCredentialsFile("credentials.json"))
	} else if apiKey := os.Getenv("GOOGLE_API_KEY"); apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	} else {
		return fmt.Errorf("credentials.json not found and GOOGLE_API_KEY not set")
	}

	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	// 1. Get Spreadsheet metadata to find all sheet titles
	resp, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve spreadsheet: %v", err)
	}

	allSheetsData := make(map[string][]map[string]interface{})

	for _, sheet := range resp.Sheets {
		title := sheet.Properties.Title

		// 2. Fetch data for each sheet
		readRange := title // Read the whole sheet
		valResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
		if err != nil {
			log.Printf("Unable to retrieve data from sheet %s: %v", title, err)
			continue
		}

		if len(valResp.Values) < 2 {
			log.Printf("Sheet %s has insufficient data", title)
			continue
		}

		// Parse rows (Assume first row is header)
		headers := valResp.Values[0]
		var sheetData []map[string]interface{}

		for _, row := range valResp.Values[1:] {
			entry := make(map[string]interface{})
			for i, cell := range row {
				if i < len(headers) {
					headerName := fmt.Sprintf("%v", headers[i])
					entry[headerName] = cell
				}
			}
			sheetData = append(sheetData, entry)
		}
		allSheetsData[title] = sheetData
	}

	jsonData, err := json.MarshalIndent(allSheetsData, "", "  ")
	if err != nil {
		return err
	}

	// Filename based on Spreadsheet Title or ID
	fileName := fmt.Sprintf("%s.json", resp.Properties.Title)
	// Sanitize filename provided it might have spaces or slashes
	fileName = strings.ReplaceAll(fileName, "/", "_")
	jsonPath := filepath.Join(jsonDir, fileName)

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("Converted Spreadsheet '%s' (%s) to %s (Sheets: %d)\n", resp.Properties.Title, spreadsheetID, jsonPath, len(allSheetsData))
	return nil
}

func convertExcelToJSON(excelPath, jsonDir string) error {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Map to hold data from ALL sheets: SheetName -> List of Rows
	allSheetsData := make(map[string][]map[string]interface{})

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return fmt.Errorf("no sheets found in %s", excelPath)
	}

	for _, sheetName := range sheets {
		rows, err := f.GetRows(sheetName)
		if err != nil {
			log.Printf("Failed to get rows for sheet %s in %s: %v", sheetName, excelPath, err)
			continue
		}

		if len(rows) < 2 {
			log.Printf("sheet %s in %s has not enough data", sheetName, excelPath)
			continue
		}

		headers := rows[0]
		var sheetData []map[string]interface{}

		for _, row := range rows[1:] {
			entry := make(map[string]interface{})
			for i, cell := range row {
				if i < len(headers) {
					entry[headers[i]] = cell
				}
			}
			sheetData = append(sheetData, entry)
		}
		allSheetsData[sheetName] = sheetData
	}

	jsonData, err := json.MarshalIndent(allSheetsData, "", "  ")
	if err != nil {
		return err
	}

	baseName := filepath.Base(excelPath)
	jsonFileName := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ".json"
	jsonPath := filepath.Join(jsonDir, jsonFileName)

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("Converted %s to %s (Sheets: %d)\n", excelPath, jsonPath, len(allSheetsData))
	return nil
}
