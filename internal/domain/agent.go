package domain

// Agent identifies which AI agent consumes a skill.
type Agent string

const (
	AgentClaude  Agent = "claude"
	AgentCopilot Agent = "copilot"
)

func (a Agent) String() string { return string(a) }
