package sessions

const WorkloadClaudeCode = "claude-code"

type Checkpoint struct {
	SizeBytes *int64 `json:"sizeBytes,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type Session struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	WorkloadType string      `json:"workloadType"`
	State        string      `json:"state"`
	Model        string      `json:"model,omitempty"`
	CreatedAt    string      `json:"createdAt"`
	LastAccess   string      `json:"lastAccess"`
	Checkpoint   *Checkpoint `json:"checkpoint,omitempty"`
}

type ReadResult struct {
	Session    *Session `json:"session,omitempty"`
	Path       string   `json:"path,omitempty"`
	Payload    string   `json:"payload"`
	NextOffset int64    `json:"nextOffset"`
}

type WriteResult struct {
	Session *Session `json:"session,omitempty"`
	Path    string   `json:"path,omitempty"`
}

type ClaudeCodeConfig struct {
	DefaultModel string   `json:"defaultModel"`
	Models       []string `json:"models"`
}

type RuntimeConfig struct {
	ClaudeCode ClaudeCodeConfig `json:"claudeCode"`
}
