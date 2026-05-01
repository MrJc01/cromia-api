package db

func (d *sqlDB) LogUsage(userID, keyID int, model string, promptTokens, completionTokens int, cost float64) error {
	_, err := d.conn.Exec(`
		INSERT INTO usage_logs (user_id, api_key_id, model_used, prompt_tokens, completion_tokens, cost_deducted)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, keyID, model, promptTokens, completionTokens, cost)
	return err
}

func (d *sqlDB) GetUserUsageLogs(userID int) ([]UsageLog, error) {
	rows, err := d.conn.Query(`
		SELECT id, user_id, api_key_id, model_used, prompt_tokens, completion_tokens, cost_deducted, created_at 
		FROM usage_logs WHERE user_id = ? ORDER BY id DESC LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []UsageLog
	for rows.Next() {
		var l UsageLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.APIKeyID, &l.ModelUsed, &l.PromptTokens, &l.CompletionTokens, &l.CostDeducted, &l.CreatedAt); err != nil {
			continue
		}
		logs = append(logs, l)
	}
	return logs, nil
}
