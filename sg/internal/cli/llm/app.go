package llm

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/Azure/ShieldGuard/sg/internal/llm"
	"github.com/Azure/ShieldGuard/sg/internal/project"
	"github.com/Azure/ShieldGuard/sg/internal/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/openai/openai-go"
	"github.com/spf13/pflag"
)

type cliApp struct {
	projectSpecFile string
	contextRoot     string

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

	llmClient, err := llm.AzureOpenAIClientFromEnv()
	if err != nil {
		return fmt.Errorf("create LLM client: %w", err)
	}

	for idx, target := range projectSpec.Files {
		if _, err := cliApp.queryFileTarget(ctx, cliApp.contextRoot, target, llmClient); err != nil {
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
		Foreground(lipgloss.Color("#FFA07A")).
		Bold(true).
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

	systemPrompt := strings.TrimSpace(target.SystemPrompt)
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
		"summary token usages",
		"completionTokens=%d, totalTokens=%d\n",
		summaryTokenUsage.CompletionTokens,
		summaryTokenUsage.TotalTokens,
	)
	pprint(
		cliApp.stdout,
		"analyze token usages",
		"completionTokens=%d, totalTokens=%d\n",
		chatCompletion.Usage.CompletionTokens,
		chatCompletion.Usage.TotalTokens,
	)

	return nil, nil
}

func (cliApp *cliApp) BindCLIFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&cliApp.projectSpecFile, "config", "c", project.SpecFileName, "Path to the project spec file.")
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

	return nil
}
