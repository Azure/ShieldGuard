package llm

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/ShieldGuard/sg/internal/llm"
	"github.com/Azure/ShieldGuard/sg/internal/llm/swarm"
	"github.com/Azure/ShieldGuard/sg/internal/project"
	"github.com/Azure/ShieldGuard/sg/internal/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/openai/openai-go"
	"github.com/spf13/pflag"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

type cliApp struct {
	projectSpecFile string
	contextRoot     string
	logsDir         string

	stdout io.Writer
}

func newCliApp(ms ...func(*cliApp)) *cliApp {
	rv := &cliApp{}

	for _, m := range ms {
		m(rv)
	}

	return rv
}

func (cliApp *cliApp) Run() error {
	if err := cliApp.defaults(); err != nil {
		return fmt.Errorf("defaults: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	projectSpec, err := project.ReadFromFile(cliApp.projectSpecFile)
	if err != nil {
		return fmt.Errorf("read project spec: %w", err)
	}
	var systemPrompt string
	if f := os.Getenv("SYSTEM_PROMPT_FILE"); f != "" {
		b, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read system prompt file from %q: %w", f, err)
		}
		systemPrompt = string(b)
	}

	llmClient, err := llm.AzureOpenAIClientFromEnv()
	if err != nil {
		return fmt.Errorf("create LLM client: %w", err)
	}

	for idx, target := range projectSpec.Files {
		if _, err := cliApp.queryFileTarget(ctx, cliApp.contextRoot, target, systemPrompt, llmClient); err != nil {
			return fmt.Errorf("run target (%s): %w", target.Name, err)
		}

		if idx > 3 {
			break
		}
	}

	return nil
}

type sourceDesc struct {
	Name    string `xml:"name"`
	Content string `xml:"content"`
}

type sourceDescs struct {
	Sources []sourceDesc `xml:"sources"`
}

func (cliApp *cliApp) queryFileTargetSummary(
	ctx context.Context,
	encodedSourceDescs string,
	llmClient *openai.Client,
) (string, openai.CompletionUsage, error) {
	systemPrompt := llm.PromptSummarizeTarget()
	userPrompt := encodedSourceDescs
	assistantPrompt := string(llm.SourceSummaryItemStartingTag)

	chatCompletion, err := llmClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
			openai.AssistantMessage(assistantPrompt),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return "", openai.CompletionUsage{}, fmt.Errorf("chat completion: %w", err)
	}
	assistantResponse := chatCompletion.Choices[0].Message.Content

	var summary llm.SourceSummary
	if err := llm.ParseResponse(
		assistantResponse,
		llm.SourceSummaryItemStartingTag,
		llm.SourceSummaryItemClosingTag,
		&summary,
	); err != nil {
		return "", openai.CompletionUsage{}, fmt.Errorf("parse assistant response: %w", err)
	}

	if summary.Content == "" {
		fmt.Printf("[DEBUG] no summary found, raw response: %s\n", assistantResponse)
	}

	return summary.Content, chatCompletion.Usage, nil
}

func pprint(out io.Writer, category string, format string, args ...interface{}) {
	category = lipgloss.NewStyle().
		// Foreground(lipgloss.Color("#FFA07A")).
		// Bold(true).
		PaddingLeft(1).
		PaddingRight(1).
		Width(152).
		Render(category)

	message := fmt.Sprintf(format, args...)
	message = lipgloss.NewStyle().Align(lipgloss.Left).
		PaddingLeft(1).
		PaddingRight(1).
		Width(152).
		Render(message)

	m := lipgloss.JoinVertical(
		lipgloss.Left,
		category, message,
	)

	m = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Render(m)

	fmt.Fprint(out, m+"\n")
}

func (cliApp *cliApp) queryFileTarget(
	ctx context.Context,
	contextRoot string,
	target project.FileTargetSpec,
	systemPrompt string,
	llmClient *openai.Client,
) ([]any, error) {
	paths := utils.Map(target.Paths, func(path string) string {
		fullPath := filepath.Join(contextRoot, path)
		fullPath = filepath.Clean(fullPath)
		return fullPath
	})

	sources, err := llm.SourcesFromPath(paths).ContextRoot(contextRoot).Complete()
	if err != nil {
		return nil, fmt.Errorf("load sources failed: %w", err)
	}

	var sourceDescs sourceDescs
	for _, source := range sources {
		content, err := source.Content()
		if err != nil {
			return nil, fmt.Errorf("read source content: %w", err)
		}

		sourceDescs.Sources = append(sourceDescs.Sources, sourceDesc{
			Name:    source.Name(),
			Content: content,
		})
	}

	encodedSourceDescs, err := xml.Marshal(sourceDescs)
	if err != nil {
		return nil, fmt.Errorf("marshal source descs: %w", err)
	}

	slog.SetLogLoggerLevel(slog.LevelDebug)
	logger := slog.Default()

	loop := swarm.New(logger, llmClient)
	resp, err := loop.Run(ctx, swarm.LoopRunParams{
		Agent:        AgentTriage,
		AgentContext: swarm.CreateAgentContext(),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(string(encodedSourceDescs)),
		},
		ExecuteTools: true,
	})
	if err != nil {
		return nil, fmt.Errorf("run loop: %w", err)
	}
	for _, message := range resp.Messages {
		mm, _ := json.Marshal(message)
		content := gjson.GetBytes(mm, "content")

		pprint(cliApp.stdout, "message", content.String())
	}

	return nil, nil
}

func (cliApp *cliApp) queryFileTarget2(
	ctx context.Context,
	contextRoot string,
	target project.FileTargetSpec,
	systemPrompt string,
	llmClient *openai.Client,
) ([]any, error) {
	paths := utils.Map(target.Paths, func(path string) string {
		fullPath := filepath.Join(contextRoot, path)
		fullPath = filepath.Clean(fullPath)
		return fullPath
	})

	sources, err := llm.SourcesFromPath(paths).ContextRoot(contextRoot).Complete()
	if err != nil {
		return nil, fmt.Errorf("load sources failed: %w", err)
	}

	var sourceDescs sourceDescs
	for _, source := range sources {
		content, err := source.Content()
		if err != nil {
			return nil, fmt.Errorf("read source content: %w", err)
		}

		sourceDescs.Sources = append(sourceDescs.Sources, sourceDesc{
			Name:    source.Name(),
			Content: content,
		})
	}

	encodedSourceDescs, err := xml.Marshal(sourceDescs)
	if err != nil {
		return nil, fmt.Errorf("marshal source descs: %w", err)
	}

	summary, summaryTokenUsage, err := cliApp.queryFileTargetSummary(
		ctx, string(encodedSourceDescs),
		llmClient,
	)
	if err != nil {
		return nil, fmt.Errorf("query file target summary: %w", err)
	}

	systemPrompt = strings.TrimSpace(systemPrompt)
	if systemPrompt == "" {
		systemPrompt = llm.PromptSummarizeTarget()
	}
	systemPrompt = llm.PromptWithOutputFormat(systemPrompt)
	systemPrompt = strings.TrimSpace(systemPrompt)

	userPrompt := string(encodedSourceDescs)
	userPrompt = llm.PromptWithSummary(userPrompt, summary)

	assistantPrompt := string(llm.AnswersItemStartingTag)

	chatCompletion, err := llmClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
			openai.AssistantMessage(assistantPrompt),
		}),
		Model: openai.F(openai.ChatModelGPT4o),
	})
	if err != nil {
		return nil, fmt.Errorf("chat completion: %w", err)
	}
	assistantResponse := chatCompletion.Choices[0].Message.Content

	pprint(cliApp.stdout, "target", target.Name)
	pprint(cliApp.stdout, "target summary", summary)
	pprint(cliApp.stdout, "system prompt", systemPrompt)

	var answers llm.AnswerItems
	if err := llm.ParseResponse(assistantResponse, llm.AnswersItemStartingTag, llm.AnswersItemClosingTag, &answers); err != nil {
		return nil, fmt.Errorf("parse assistant response: %w", err)
	}

	for _, answer := range answers.Items {
		pprint(cliApp.stdout, "answer", answer.String())
	}

	pprint(
		cliApp.stdout,
		"summary token usage",
		"completionTokens=%d, totalTokens=%d\n",
		summaryTokenUsage.CompletionTokens,
		summaryTokenUsage.TotalTokens,
	)
	pprint(
		cliApp.stdout,
		"analyze token usage",
		"completionTokens=%d, totalTokens=%d\n",
		chatCompletion.Usage.CompletionTokens,
		chatCompletion.Usage.TotalTokens,
	)

	if err := cliApp.logQueryFileTarget(
		ctx,
		target,
		systemPrompt, userPrompt, summary, assistantResponse,
		summaryTokenUsage, chatCompletion.Usage,
	); err != nil {
		return nil, fmt.Errorf("log query file target: %w", err)
	}

	return nil, nil
}

type queryFileTarget struct {
	Target  project.FileTargetSpec `json:"target"`
	Prompts struct {
		System string `json:"system"`
		User   string `json:"user"`
	} `json:"prompts"`
	Responses struct {
		Summary string `json:"summary"`
		Analyze string `json:"assistant"`
	} `json:"responses"`
	Usages struct {
		Summary openai.CompletionUsage `json:"summary"`
		Analyze openai.CompletionUsage `json:"analyze"`
	} `json:"usages"`
}

func (cliApp *cliApp) logQueryFileTarget(
	ctx context.Context,
	target project.FileTargetSpec,
	systemPrompt string,
	userPrompt string,
	summaryResponse string,
	analyzeResponse string,
	summaryTokenUsage openai.CompletionUsage,
	analyzeTokenUsage openai.CompletionUsage,
) error {
	d := queryFileTarget{
		Target: target,
	}
	d.Prompts.System = systemPrompt
	d.Prompts.User = userPrompt
	d.Responses.Summary = summaryResponse
	d.Responses.Analyze = analyzeResponse
	d.Usages.Summary = summaryTokenUsage
	d.Usages.Analyze = analyzeTokenUsage

	outputFile := filepath.Join(cliApp.logsDir, fmt.Sprintf("%s.yaml", target.Name))
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer output.Close()

	enc := yaml.NewEncoder(output)
	enc.SetIndent(2)
	if err := enc.Encode(d); err != nil {
		return fmt.Errorf("encode output: %w", err)
	}

	return nil
}

func (cliApp *cliApp) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&cliApp.projectSpecFile, "config", "c", project.SpecFileName, "Path to the project spec file.")
	fs.StringVar(&cliApp.logsDir, "logs-dir", "", "Logs dir to use. Defaults to the $pwd/logs.")
}

func (cliApp *cliApp) defaults() error {
	var err error

	if cliApp.projectSpecFile == "" {
		return fmt.Errorf("project spec file is not specified")
	}
	cliApp.projectSpecFile, err = filepath.Abs(cliApp.projectSpecFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of the project spec file: %w", err)
	}

	if cliApp.contextRoot == "" {
		cliApp.contextRoot = "."
	}
	cliApp.contextRoot, err = filepath.Abs(cliApp.contextRoot)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of the context root: %w", err)
	}

	if cliApp.logsDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("get current working directory: %w", err)
		}

		logDir := time.Now().Format("2006-01-02-15-04-05")

		cliApp.logsDir = filepath.Join(cwd, "logs", logDir)
	}
	cliApp.logsDir, err = filepath.Abs(cliApp.logsDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path of the logs dir: %w", err)
	}
	if err := os.MkdirAll(cliApp.logsDir, 0755); err != nil {
		return fmt.Errorf("create logs dir: %w", err)
	}

	if cliApp.stdout == nil {
		outputLogsFile := filepath.Join(cliApp.logsDir, "output.log")
		outputLogs, err := os.Create(outputLogsFile)
		if err != nil {
			return fmt.Errorf("create output logs file: %w", err)
		}
		// FIXME: outputLogs should be closed at process exit
		cliApp.stdout = io.MultiWriter(os.Stdout, outputLogs)
	}

	return nil
}
