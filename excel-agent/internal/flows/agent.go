package flows

import (
	"context"
	"fmt"

	"excel-agent/internal/config"
	"excel-agent/internal/processor"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

func registerTools(g *genkit.Genkit, cfg *config.Config, registry map[string]interface{}) ai.Tool {
	// Register Redis Query Tool
	redisTool := genkit.DefineTool(
		g,
		"queryRedis",
		"Queries spreadsheet data from Redis using a key. Key format is usually 'FileName:SheetName'.",
		func(ctx *ai.ToolContext, input *processor.RedisQueryInput) (*processor.RedisQueryOutput, error) {
			return processor.QueryRedisTool(ctx, input, cfg.RedisAddr, cfg.RedisDB)
		},
	)
	registry["queryRedis"] = redisTool
	return redisTool
}

func registerAgentFlows(g *genkit.Genkit, cfg *config.Config, registry map[string]interface{}, redisTool ai.Tool) {
	// Define the Smart Query Flow (Agent)
	registry["queryFlow"] = genkit.DefineFlow(g, "queryFlow", func(ctx context.Context, prompt string) (string, error) {
		systemPrompt := `You are an assistant that analyzes spreadsheet data stored in Redis.
You can use the 'queryRedis' tool to fetch data. 
The keys for the tool are in the format 'FileName:SheetName'.
Available data files include: Arena, BaseOption, BattleGroup, Building, Character, Dialogue, Dungeon, Gacha, Item, LocalSeet, Node, Reward, Shop, Sound, Stat, Tag, Tutorial, Upgrade, WorldMap.
When a user asks for data, first determine the correct key, fetch the data, and then provide a concise summary or answer based on the retrieved JSON.
If the JSON is too large, summarize the most relevant parts.
`
		resp, err := genkit.GenerateText(ctx, g,
			ai.WithSystem(systemPrompt),
			ai.WithPrompt(prompt),
			ai.WithTools(redisTool),
		)
		if err != nil {
			return "", fmt.Errorf("AI agent query failed: %v", err)
		}
		return resp, nil
	})
}
