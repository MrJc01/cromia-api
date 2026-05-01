package db

type User struct {
	ID           int
	Username     string
	PasswordHash string
	Balance      float64
	CreatedAt    string
}

type APIKey struct {
	ID        int
	UserID    int
	Name      string
	KeyHash   string
	CreatedAt string
	RevokedAt *string
}

type ProviderModel struct {
	ID             int
	ProviderName   string
	ModelName      string
	CostMultiplier float64
	PromptCost     float64
	CompletionCost float64
	IsActive       bool
}

type UsageLog struct {
	ID               int     `json:"id"`
	UserID           int     `json:"user_id"`
	APIKeyID         int     `json:"api_key_id"`
	ModelUsed        string  `json:"model_used"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CostDeducted     float64 `json:"cost_deducted"`
	CreatedAt        string  `json:"created_at"`
}



type DB interface {
	// Users
	CreateUser(username, passwordHash string, initialBalance float64) (int, error)
	GetUserByUsername(username string) (*User, error)
	GetUserByID(id int) (*User, error)
	AddBalance(userID int, amount float64) error
	DeductBalance(userID int, amount float64) error
	ListUsers() ([]User, error)

	// API Keys
	CreateKey(userID int, name, keyHash string) (int, error)
	GetActiveKeys() ([]APIKey, error)
	RevokeKey(id int) error

	// Provider Models
	EnableModel(provider, model string, multiplier float64) error
	DisableModel(model string) error
	GetActiveModels() ([]ProviderModel, error)
	UpdateModelPricing(modelName string, promptCost, completionCost float64) error

	// Usage Logging
	LogUsage(userID, keyID int, model string, promptTokens, completionTokens int, cost float64) error
	GetUserUsageLogs(userID int) ([]UsageLog, error)
}
