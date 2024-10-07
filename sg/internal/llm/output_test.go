package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseOutput_Summary(t *testing.T) {
	s := "foo</summary>"

	var dest SourceSummary
	err := ParseResponse(s, SourceSummaryItemStartingTag, SourceSummaryItemClosingTag, &dest)
	assert.NoError(t, err)
	assert.Equal(t, "foo", dest.Content)
}
