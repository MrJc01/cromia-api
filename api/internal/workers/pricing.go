package workers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"cromia/api/internal/db"
)

type ORModelPricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

type ORModel struct {
	ID      string         `json:"id"`
	Pricing ORModelPricing `json:"pricing"`
}

type ORResponse struct {
	Data []ORModel `json:"data"`
}

// SyncPricing fetches the latest pricing from OpenRouter and updates the local DB
func SyncPricing(database db.DB) error {
	log.Println("[PricingWorker] Fetching pricing from OpenRouter API...")
	
	openRouterURL := os.Getenv("OPENROUTER_BASE_URL")
	if openRouterURL == "" {
		openRouterURL = "https://openrouter.ai/api/v1"
	}

	resp, err := http.Get(openRouterURL + "/models")
	if err != nil {
		return fmt.Errorf("failed to fetch from openrouter: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openrouter returned status: %d", resp.StatusCode)
	}

	var payload ORResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("failed to decode openrouter response: %w", err)
	}

	activeModels, err := database.GetActiveModels()
	if err != nil {
		return fmt.Errorf("failed to get active models: %w", err)
	}

	// Create a map of active models to update
	activeMap := make(map[string]db.ProviderModel)
	for _, m := range activeModels {
		activeMap[m.ModelName] = m
	}

	updatedCount := 0

	for _, orModel := range payload.Data {
		// Map OpenRouter ID to our ModelName. 
		// Example: "deepseek/deepseek-chat" in OR might map to "deepseek-chat" in our system.
		// "openai/gpt-3.5-turbo" might map to "gpt-3.5-turbo" or "openai/gpt-3.5-turbo".
		
		internalName := orModel.ID
		if strings.Contains(orModel.ID, "/") {
			parts := strings.Split(orModel.ID, "/")
			internalName = parts[1] // Extract "deepseek-chat" from "deepseek/deepseek-chat"
		}

		if _, exists := activeMap[internalName]; exists {
			promptCost, _ := strconv.ParseFloat(orModel.Pricing.Prompt, 64)
			compCost, _ := strconv.ParseFloat(orModel.Pricing.Completion, 64)

			if err := database.UpdateModelPricing(internalName, promptCost, compCost); err != nil {
				log.Printf("[PricingWorker] Error updating %s: %v", internalName, err)
			} else {
				log.Printf("[PricingWorker] Synced %s: Prompt=$%f, Completion=$%f", internalName, promptCost, compCost)
				updatedCount++
			}
			delete(activeMap, internalName)
		} else if _, exists := activeMap[orModel.ID]; exists {
			// Exact match (e.g., if user enabled "deepseek/deepseek-chat" literally)
			promptCost, _ := strconv.ParseFloat(orModel.Pricing.Prompt, 64)
			compCost, _ := strconv.ParseFloat(orModel.Pricing.Completion, 64)

			if err := database.UpdateModelPricing(orModel.ID, promptCost, compCost); err != nil {
				log.Printf("[PricingWorker] Error updating %s: %v", orModel.ID, err)
			} else {
				log.Printf("[PricingWorker] Synced %s: Prompt=$%f, Completion=$%f", orModel.ID, promptCost, compCost)
				updatedCount++
			}
			delete(activeMap, orModel.ID)
		}
	}

	log.Printf("[PricingWorker] Sync complete. Updated %d models.", updatedCount)
	return nil
}
