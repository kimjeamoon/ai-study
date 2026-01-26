package processor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

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
