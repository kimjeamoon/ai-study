package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// ProcessXlsxFiles converts all .xlsx files in a directory to JSON.
func ProcessXlsxFiles(xlsxDir, jsonDir string) (int, error) {
	files, err := os.ReadDir(xlsxDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read xlsx directory: %w", err)
	}

	processedCount := 0
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".xlsx" {
			continue
		}

		filePath := filepath.Join(xlsxDir, file.Name())
		if err := ConvertExcelToJSON(filePath, jsonDir); err != nil {
			log.Printf("Failed to convert %s: %v", file.Name(), err)
			continue
		}
		processedCount++
	}

	return processedCount, nil
}

// ConvertExcelToJSON converts a single Excel file to JSON.
func ConvertExcelToJSON(excelPath, jsonDir string) error {
	f, err := excelize.OpenFile(excelPath)
	if err != nil {
		return err
	}
	defer f.Close()

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

// ConvertGoogleSheetToJSON fetches data from a Google Spreadsheet and saves it as JSON.
func ConvertGoogleSheetToJSON(ctx context.Context, spreadsheetID, jsonDir, apiKey string) error {
	var opts []option.ClientOption

	if _, err := os.Stat("credentials.json"); err == nil {
		opts = append(opts, option.WithCredentialsFile("credentials.json"))
	} else if apiKey != "" {
		opts = append(opts, option.WithAPIKey(apiKey))
	} else {
		return fmt.Errorf("credentials.json not found and GOOGLE_API_KEY not set")
	}

	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	resp, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve spreadsheet: %v", err)
	}

	allSheetsData := make(map[string][]map[string]interface{})

	for _, sheet := range resp.Sheets {
		title := sheet.Properties.Title

		valResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, title).Do()
		if err != nil {
			log.Printf("Unable to retrieve data from sheet %s: %v", title, err)
			continue
		}

		if len(valResp.Values) < 2 {
			log.Printf("Sheet %s has insufficient data", title)
			continue
		}

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

	fileName := fmt.Sprintf("%s.json", resp.Properties.Title)
	fileName = strings.ReplaceAll(fileName, "/", "_")
	jsonPath := filepath.Join(jsonDir, fileName)

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return err
	}

	fmt.Printf("Converted Spreadsheet '%s' (%s) to %s (Sheets: %d)\n", resp.Properties.Title, spreadsheetID, jsonPath, len(allSheetsData))
	return nil
}
