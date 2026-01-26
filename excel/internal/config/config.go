package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	XlsxDir       string
	JsonDir       string
	DataDir       string
	DefaultModel  string
	GoogleSheetID string
	GoogleAPIKey  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		XlsxDir:       getEnv("XLSX_DIR", "xlsx"),
		JsonDir:       getEnv("JSON_DIR", "json"),
		DataDir:       getEnv("DATA_DIR", "data"),
		DefaultModel:  getEnv("DEFAULT_MODEL", "googleai/gemini-2.5-flash"),
		GoogleSheetID: os.Getenv("GOOGLE_SHEET_ID"),
		GoogleAPIKey:  os.Getenv("GOOGLE_API_KEY"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func (c *Config) EnsureDirs() error {
	dirs := []string{c.XlsxDir, c.JsonDir, c.DataDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
