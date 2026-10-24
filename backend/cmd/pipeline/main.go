package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"meta-targeting-copilot/backend/internal/ai"
	"meta-targeting-copilot/backend/internal/config"
	"meta-targeting-copilot/backend/internal/domain"
	"meta-targeting-copilot/backend/internal/meta"
	"meta-targeting-copilot/backend/internal/pipeline"
	"meta-targeting-copilot/backend/internal/strategy"
	"meta-targeting-copilot/backend/internal/targeting"
)

func main() {
	product := flag.String("product", "Fitness club", "product / service")
	country := flag.String("country", "Kazakhstan", "country")
	city := flag.String("city", "Astana", "city")
	audience := flag.String("audience", "People who already exercise regularly.", "audience prompt")
	md := flag.Bool("md", false, "print markdown instead of JSON")
	flag.Parse()

	cfg := config.Load()
	var llm ai.Provider = ai.NewMock()
	if cfg.LLMAPIKey != "" {
		llm = ai.NewOpenAI(cfg.LLMAPIKey, cfg.LLMBaseURL, cfg.LLMModel)
	}
	var metaProv meta.Provider = meta.NewMock()
	if cfg.MetaAccessToken != "" {
		metaProv = meta.NewGraph(cfg.MetaAccessToken, cfg.MetaAdAccountID, cfg.MetaAPIVersion)
	}

	in := domain.BriefInput{
		Product: *product, AudiencePrompt: *audience, Location: *city + ", " + *country,
		Country: *country, City: *city, Objective: "leads",
	}
	res, err := pipeline.New(ai.NewEngine(llm), targeting.NewResolver(metaProv)).Run(context.Background(), in, func(stage domain.JobStage) {
		fmt.Fprintf(os.Stderr, "stage %s\n", stage)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *md {
		fmt.Print(strategy.Markdown(res))
		return
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(res)
}
