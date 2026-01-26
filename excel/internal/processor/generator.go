package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func GenerateStructs(ctx context.Context, g *genkit.Genkit, fileName, jsonDir, dataDir string) (string, error) {
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

	var jsonRaw map[string]interface{}
	if err := json.Unmarshal(data, &jsonRaw); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

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
	if err := os.WriteFile(filepath.Join(dataDir, goFileName), []byte(code), 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %v", goFileName, err)
	}

	return "Successfully generated " + goFileName + " using AI", nil
}
