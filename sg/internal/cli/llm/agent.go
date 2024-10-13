package llm

import (
	"github.com/Azure/ShieldGuard/sg/internal/llm/swarm"
	"github.com/openai/openai-go"
)

var AgentSummary = swarm.AgentDecl{
	Name:              "summary",
	Description:       "Summarize the input text.",
	Model:             openai.ChatModelGPT4o,
	Instructions:      swarm.AgentInstructions(""),
	ToolChoice:        openai.ChatCompletionToolChoiceOptionStringNone,
	ParallelToolsCall: false,
}.Build()

var summary = swarm.AgentToFunction(AgentSummary)

var AgentAnalyzeServiceAccountUsage = swarm.AgentDecl{
	Name:              "analyze-service-account-usage",
	Description:       `Analyze the usage of a service account.`,
	Model:             openai.ChatModelGPT4o,
	ToolChoice:        openai.ChatCompletionToolChoiceOptionStringNone,
	Instructions:      swarm.AgentInstructions(`Analyze the usage of a service account. Report unexpected usages for both service account and RBAC settings.`),
	ParallelToolsCall: false,
}.Build()

var analyzeServiceAccountUsage = swarm.AgentToFunction(AgentAnalyzeServiceAccountUsage)

var AgentTriage = swarm.AgentDecl{
	Name:         "triage",
	Description:  `Triage the input text.`,
	Model:        openai.ChatModelGPT4o,
	Instructions: swarm.AgentInstructions("Read and understand the input text. React based on the content."),
	ToolChoice:   openai.ChatCompletionToolChoiceOptionStringAuto,
	Functions: []swarm.AgentFunction{
		summary,
		analyzeServiceAccountUsage,
	},
	ParallelToolsCall: true,
}.Build()
