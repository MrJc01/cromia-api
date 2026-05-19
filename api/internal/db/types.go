package db

type User struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	PasswordHash string  `json:"-"`
	Balance      float64 `json:"balance"`
	CreatedAt    string  `json:"created_at"`
}

type APIKey struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	Name      string  `json:"name"`
	KeyHash   string  `json:"key_hash"`
	CreatedAt string  `json:"created_at"`
	RevokedAt *string `json:"revoked_at"`
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
	UpdatePassword(userID int, newPasswordHash string) error
	AddBalance(userID int, amount float64) error
	DeductBalance(userID int, amount float64) error
	ListUsers() ([]User, error)

	// API Keys
	CreateKey(userID int, name, keyHash string) (int, error)
	GetActiveKeys() ([]APIKey, error)
	GetUserKeys(userID int) ([]APIKey, error)
	GetKeyByHash(keyHash string) (*APIKey, error)
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
