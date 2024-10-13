package llm

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Source interface {
	// Name returns the name of the target.
	Name() string

	// Content returns the content of the target.
	Content() (string, error)
}

type Answer struct {
	SourceLocation string `xml:"sourceLocation"`
	Comment        string `xml:"comment"`
	Suggestion     string `xml:"suggestion"`
	Thinking       string `xml:"thinking"`
	Reflection     string `xml:"reflection"`
}

func (s Answer) String() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "SourceLocation: %s\n", s.SourceLocation)
	fmt.Fprintf(&sb, "Comment: %s\n", s.Comment)
	fmt.Fprintf(&sb, "Suggestion:\n  %s\n\n", s.Suggestion)
	fmt.Fprintf(&sb, "Thinking:\n  %s\n\n", s.Thinking)
	fmt.Fprintf(&sb, "Reflection: %s", s.Reflection)

	return sb.String()
}

type AnswerItems struct {
	Items []Answer `xml:"answer"`
}

const (
	AnswersItemStartingTag = "<answers>"
	AnswersItemClosingTag  = "</answers>"
)

type SourceSummary struct {
	XMLName xml.Name `xml:"summary"`
	Content string   `xml:",chardata"`
}

const (
	SourceSummaryItemStartingTag = "<summary>"
	SourceSummaryItemClosingTag  = "</summary>"
)
