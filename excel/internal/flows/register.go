package flows

import (
	"excel-agent/internal/config"

	"github.com/firebase/genkit/go/genkit"
)

// RegisterFlows initializes and registers all tools and flows in the project.
func RegisterFlows(g *genkit.Genkit, cfg *config.Config) map[string]interface{} {
	registry := make(map[string]interface{})

	// 1. Register Tools & Local Logic
	redisTool := registerTools(g, cfg, registry)

	// 2. Register Processing (Conversion) Flows
	registerProcessingFlows(g, cfg, registry)

	// 3. Register AI-driven Flows
	registerGeneratorFlows(g, cfg, registry)
	registerAgentFlows(g, cfg, registry, redisTool)

	return registry
}
