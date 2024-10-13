package llm

import (
	"fmt"
	"strings"

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

var analyzeServicePort = swarm.ToAgentFunctionOrPanic(
	"check-port",
	"Analyze the exposed port for a service. The first argument is the k8s service type (LoadBalancer / ClusterIP / NodePort etc) and the second argument is the port number. Invoke more times to check multiple ports.",
	func(svcType string, port int64) (string, error) {
		if strings.ToLower(svcType) == "clusterip" {
			return "Service type is ClusterIP. No need to check exposed ports.", nil
		}

		if port == 80 {
			return "Port 80 is exposed. Consider using HTTPS instead.", nil
		}
		if port == 443 {
			return "Port 443 is exposed. Make sure it is secure.", nil
		}
		return fmt.Sprintf("Port %d is exposed. Check if it is necessary.", port), nil
	},
)

var AgentAnalyzeServicePort = swarm.AgentDecl{
	Name:              "analyze-service-ports",
	Description:       `Analyze the exposed ports for a service.`,
	Model:             openai.ChatModelGPT4o,
	Instructions:      swarm.AgentInstructions(`Analyze the exposed ports from the declared spec. Inspect evert service and port combination. Report any potential security risks.`),
	ToolChoice:        openai.ChatCompletionToolChoiceOptionStringAuto,
	ParallelToolsCall: true,
	Functions: []swarm.AgentFunction{
		analyzeServicePort,
	},
}.Build()

var AgentTriage = swarm.AgentDecl{
	Name:         "triage",
	Description:  `Triage the input text.`,
	Model:        openai.ChatModelGPT4o,
	Instructions: swarm.AgentInstructions("Read and understand the input text. React based on the suer input content. You can perform multiple analysis in parallel. Summarize your findings."),
	ToolChoice:   openai.ChatCompletionToolChoiceOptionStringAuto,
	Functions: []swarm.AgentFunction{
		summary,
		analyzeServiceAccountUsage,
		swarm.AgentToFunction(AgentAnalyzeServicePort),
	},
	ParallelToolsCall: true,
}.Build()
