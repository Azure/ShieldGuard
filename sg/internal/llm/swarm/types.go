package swarm

import (
	"context"
	"encoding/json"

	"github.com/openai/openai-go"
)

// TODO: declare as interface...
type AgentContext struct {
	Parameters map[string]any
}

func CreateAgentContext() AgentContext {
	return AgentContext{Parameters: make(map[string]any)}
}

func (ac AgentContext) Clone() AgentContext {
	newParams := make(map[string]any)
	for k, v := range ac.Parameters {
		newParams[k] = v
	}
	return AgentContext{Parameters: newParams}
}

func (ac AgentContext) Merge(other AgentContext) AgentContext {
	rv := ac.Clone()
	for k, v := range other.Parameters {
		rv.Parameters[k] = v
	}
	return rv
}

type AgentResult struct {
	Value        string
	Agent        Agent
	AgentContext AgentContext
}

type AgentFunction interface {
	ToolParam() openai.ChatCompletionToolParam
	Invoke(
		ctx context.Context,
		agentCtx AgentContext,
		arguments json.RawMessage,
	) (*AgentResult, error)
}

type Agent interface {
	agentImpl()

	Name() string
	Description() string
	Model() string
	Instructions(agentCtx AgentContext) string
	Functions() []AgentFunction
	ToolChoice() openai.ChatCompletionToolChoiceOptionUnionParam
	ParallelToolsCall() bool
}

type Response struct {
	Messages     []openai.ChatCompletionMessageParamUnion
	Agent        Agent
	AgentContext AgentContext
}
