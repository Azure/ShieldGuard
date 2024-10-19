package llm

import (
	"fmt"
	"strings"

	"github.com/b4fun/swarmctl/swarm"
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

type paramAnalyzeServicePort struct {
	SvcType string `json:"svcType" json_desc:"The type of the k8s service."`
	Port    int64  `json:"port" json_desc:"The exposed port of the service."`
}

var analyzeServicePort = swarm.ToAgentFunctionOrPanic(
	"check-port",
	"Analyze the exposed port for a service.",
	func(param paramAnalyzeServicePort) (string, error) {
		if strings.ToLower(param.SvcType) == "clusterip" {
			return "Service type is ClusterIP. No need to check exposed ports.", nil
		}

		if param.Port == 80 {
			return "Port 80 is exposed. Consider using HTTPS instead.", nil
		}
		if param.Port == 443 {
			return "Port 443 is exposed. Make sure it is secure.", nil
		}
		return fmt.Sprintf("Port %d is exposed. Check if it is necessary.", param.Port), nil
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
	},
	ParallelToolsCall: true,
}.Build()

func init() {
	triage := swarm.AgentToFunction(AgentTriage)

	AgentAnalyzeServicePort.HackAddFunction(triage)
	AgentAnalyzeServiceAccountUsage.HackAddFunction(triage)
}
