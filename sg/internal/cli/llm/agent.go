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

var testFunction = swarm.ToAgentFunctionOrPanic(
	"test",
	"Verify the output. The first argument is the input text to summarize. This function should be called as the last function.",
	func(input string) (string, error) {
		return input, nil
	},
)

var AgentTriage = swarm.AgentDecl{
	Name:        "triage",
	Description: `Triage the input text.`,
	Model:       openai.ChatModelGPT4o,
	Instructions: swarm.AgentInstructions(`
Read and understand the input text. React based on the suer input content. You can perform multiple analysis in parallel.

Summarize the output from multiple analysis. Put output as XML format under the <feedbacks>. Each analysis output should be a child element enclosed with <answer> tag.
Each answer should contain the following information:

- <analyze-type> tag: the type of analysis
- <result> tag: the result of the analysis
- <suggestion> tag: the suggestion based on the analysis result
- <comment> tag: any additional comments
- <thinking> tag: the thought process of the analysis
- <reflection> tag: provide a honest reflection of your feedback and reflection process by rating yourself on a scale of 1-10 (1 is least and 10 is most). Provide rating only.

Make sure your final output is well-formatted XML. Run the "test" function call as the last before submitting the final output.

Call this agent for general questions and when no other agent is correct for the user query.
`),
	ToolChoice: openai.ChatCompletionToolChoiceOptionStringAuto,
	Functions: []swarm.AgentFunction{
		summary,
		analyzeServiceAccountUsage,
		swarm.AgentToFunction(AgentAnalyzeServicePort),
		testFunction,
	},
	ParallelToolsCall: true,
}.Build()
