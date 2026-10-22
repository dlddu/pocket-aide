package sessions

const (
	WorkloadClaudeCode  = "claude-code"
	InteractionAgent    = "agent"
	InteractionTerminal = "terminal"
)

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

type WorkloadTypeInfo struct {
	Type        string            `json:"type"`
	Label       string            `json:"label,omitempty"`
	Description string            `json:"description,omitempty"`
	Interaction string            `json:"interaction"`
	Model       *ClaudeCodeConfig `json:"model,omitempty"`
}

type RuntimeConfig struct {
	ClaudeCode    ClaudeCodeConfig   `json:"claudeCode"`
	WorkloadTypes []WorkloadTypeInfo `json:"workloadTypes"`
}

func (c *RuntimeConfig) Normalize() {
	if c.ClaudeCode.Models == nil {
		c.ClaudeCode.Models = []string{}
	}
	if len(c.WorkloadTypes) == 0 {
		legacy := c.ClaudeCode
		c.WorkloadTypes = []WorkloadTypeInfo{{Type: WorkloadClaudeCode, Interaction: InteractionAgent, Model: &legacy}}
	}
	for i := range c.WorkloadTypes {
		if c.WorkloadTypes[i].Interaction != InteractionTerminal {
			c.WorkloadTypes[i].Interaction = InteractionAgent
		}
		if m := c.WorkloadTypes[i].Model; m != nil && m.Models == nil {
			m.Models = []string{}
		}
	}
}

func (c RuntimeConfig) Workload(t string) (WorkloadTypeInfo, bool) {
	for _, w := range c.WorkloadTypes {
		if w.Type == t {
			return w, true
		}
	}
	return WorkloadTypeInfo{}, false
}
