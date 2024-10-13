package swarm

import "github.com/openai/openai-go"

type AgentDecl struct {
	Name              string
	Description       string
	Model             string
	Instructions      func(agentCtx AgentContext) string
	Functions         []AgentFunction
	ToolChoice        openai.ChatCompletionToolChoiceOptionUnionParam
	ParallelToolsCall bool
}

func AgentInstructions(s string) func(agentCtx AgentContext) string {
	return func(agentCtx AgentContext) string {
		return s
	}
}

func (d AgentDecl) Build() Agent {
	return &agentImpl{decl: d}
}

type agentImpl struct {
	decl AgentDecl
}

var _ Agent = (*agentImpl)(nil)

func (a *agentImpl) agentImpl() {}

func (a *agentImpl) Name() string {
	return a.decl.Name
}

func (a *agentImpl) Description() string {
	return a.decl.Description
}

func (a *agentImpl) Model() string {
	return a.decl.Model
}

func (a *agentImpl) Instructions(agentCtx AgentContext) string {
	return a.decl.Instructions(agentCtx)
}

func (a *agentImpl) Functions() []AgentFunction {
	return a.decl.Functions
}

func (a *agentImpl) ToolChoice() openai.ChatCompletionToolChoiceOptionUnionParam {
	return a.decl.ToolChoice
}

func (a *agentImpl) ParallelToolsCall() bool {
	return a.decl.ParallelToolsCall
}
