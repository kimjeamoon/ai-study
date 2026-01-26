package flows

import (
	"context"

	"excel-agent/internal/config"
	"excel-agent/internal/processor"

	"github.com/firebase/genkit/go/genkit"
)

func registerGeneratorFlows(g *genkit.Genkit, cfg *config.Config, registry map[string]interface{}) {
	// AI Go Struct Generator Flow
	registry["generateStructsFlow"] = genkit.DefineFlow(g, "generateStructsFlow", func(ctx context.Context, fileName string) (string, error) {
		return processor.GenerateStructs(ctx, g, fileName, cfg.JsonDir, cfg.DataDir)
	})
}
