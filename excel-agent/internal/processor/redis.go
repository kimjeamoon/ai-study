package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/redis/go-redis/v9"
)

type RedisQueryInput struct {
	Key string `json:"key" description:"The Redis key to query (e.g., 'Arena:ArenaRankingBot')"`
}

type RedisQueryOutput struct {
	Data string `json:"data"`
}

func CacheJSONToRedis(ctx context.Context, jsonDir, redisAddr string, redisDB int) error {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDB,
	})
	defer rdb.Close()

	// Check connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	files, err := os.ReadDir(jsonDir)
	if err != nil {
		return fmt.Errorf("failed to read json directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(jsonDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read %s: %v", file.Name(), err)
			continue
		}

		var jsonRaw map[string]interface{}
		if err := json.Unmarshal(data, &jsonRaw); err != nil {
			log.Printf("Failed to parse %s: %v", file.Name(), err)
			continue
		}

		baseName := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))

		for sheetName, content := range jsonRaw {
			rows, ok := content.([]interface{})
			if !ok {
				continue
			}

			// Key format: FileName:SheetName
			key := fmt.Sprintf("%s:%s", baseName, sheetName)

			// Store as JSON string in Redis
			jsonData, err := json.Marshal(rows[1:])
			if err != nil {
				log.Printf("Failed to marshal data for key %s: %v", key, err)
				continue
			}

			if err := rdb.Set(ctx, key, jsonData, 0).Err(); err != nil {
				log.Printf("Failed to set key %s in redis: %v", key, err)
				continue
			}
			log.Printf("Cached key to Redis: %s", key)
		}
	}

	return nil
}

func GetDataFromRedis(ctx context.Context, key, redisAddr string, redisDB int) (string, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   redisDB,
	})
	defer rdb.Close()

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key '%s' not found", key)
	} else if err != nil {
		return "", err
	}

	return val, nil
}

func QueryRedisTool(ctx *ai.ToolContext, input *RedisQueryInput, redisAddr string, redisDB int) (*RedisQueryOutput, error) {
	val, err := GetDataFromRedis(ctx, input.Key, redisAddr, redisDB)
	if err != nil {
		return &RedisQueryOutput{Data: err.Error()}, nil
	}
	return &RedisQueryOutput{Data: val}, nil
}
