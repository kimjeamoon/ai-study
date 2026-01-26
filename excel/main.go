package main

import (
	"context"
	"log"

	"excel-agent/internal/cmd"
	"excel-agent/internal/config"
	"excel-agent/internal/flows"

	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
)

func main() {
	c := cmd.ParseFlags()
	cfg := config.LoadConfig()
	if err := cfg.EnsureDirs(); err != nil {
		log.Fatalf("Failed to ensure directories: %v", err)
	}

	ctx := context.Background()

	// Init Genkit with Google AI plugin
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel(cfg.DefaultModel),
	)

	// Register all flows for Genkit agent mode
	allFlows := flows.RegisterFlows(g, cfg)

	// Handle CLI commands if provided
	if c.Handle(ctx, g, cfg, allFlows) {
		return
	}

	log.Println("Genkit agent started. Waiting for requests...")
	// Block until context is cancelled
	<-ctx.Done()
}
