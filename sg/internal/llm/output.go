package llm

import (
	"encoding/xml"
	"strings"
)

func ParseResponse(
	response string,
	optionalStartingTags string,
	optionalClosingTags string,
	dest any,
) error {
	response = strings.TrimSpace(response)
	if !strings.HasPrefix(response, optionalStartingTags) {
		response = optionalStartingTags + response
	}
	if !strings.HasSuffix(response, optionalClosingTags) {
		response = response + optionalClosingTags
	}

	dec := xml.NewDecoder(strings.NewReader(response))
	dec.Strict = false // LLM's content might be XML escaped, so we need to be lenient.

	return dec.Decode(dest)
}
