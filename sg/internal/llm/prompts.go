package llm

import (
	_ "embed"
)

//go:embed prompt-output-format.txt
var promptOutputFormat []byte

func PromptWithOutputFormat(b string) string {
	return b + "\n" + string(promptOutputFormat)
}

//go:embed prompt-summarize-target.txt
var promptSummarizeTarget []byte

func PromptSummarizeTarget() string {
	return string(promptSummarizeTarget)
}

func PromptWithSummary(b string, summary string) string {
	return b + SourceSummaryItemStartingTag + summary + SourceSummaryItemClosingTag
}

//go:embed prompt-default-system-prompt.txt
var promptDefaultSystemPrompt []byte

func PromptDefaultSystem() string {
	return string(promptDefaultSystemPrompt)
}
