package db

import "database/sql"

func (d *sqlDB) EnableModel(provider, model string, multiplier float64) error {
	_, err := d.conn.Exec(`
		INSERT INTO provider_models (provider_name, model_name, cost_multiplier, is_active)
		VALUES (?, ?, ?, 1)
		ON CONFLICT(model_name) DO UPDATE SET 
			provider_name=excluded.provider_name,
			cost_multiplier=excluded.cost_multiplier,
			is_active=1
	`, provider, model, multiplier)
	return err
}

func (d *sqlDB) DisableModel(model string) error {
	_, err := d.conn.Exec("UPDATE provider_models SET is_active = 0 WHERE model_name = ?", model)
	return err
}

func (d *sqlDB) GetActiveModels() ([]ProviderModel, error) {
	rows, err := d.conn.Query(`
		SELECT id, provider_name, model_name, cost_multiplier, prompt_cost, completion_cost, is_active 
		FROM provider_models WHERE is_active = 1
	`)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	var models []ProviderModel
	for rows.Next() {
		var pm ProviderModel
		if err := rows.Scan(&pm.ID, &pm.ProviderName, &pm.ModelName, &pm.CostMultiplier, &pm.PromptCost, &pm.CompletionCost, &pm.IsActive); err != nil {
			continue
		}
		models = append(models, pm)
	}
	return models, nil
}

func (d *sqlDB) UpdateModelPricing(modelName string, promptCost, completionCost float64) error {
	_, err := d.conn.Exec(`
		UPDATE provider_models 
		SET prompt_cost = ?, completion_cost = ?
		WHERE model_name = ?
	`, promptCost, completionCost, modelName)
	return err
}

