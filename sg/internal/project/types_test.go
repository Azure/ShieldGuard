package project

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_strListOrMap(t *testing.T) {
	cases := []struct {
		input          string
		expectedValues []string
		expectErr      bool
	}{
		{
			input:          `["a", "b", "c"]`,
			expectedValues: []string{"a", "b", "c"},
		},
		{
			input:          `["a", "c", "b"]`,
			expectedValues: []string{"a", "b", "c"},
		},
		{
			input:          `{"a": null, "b": null, "c": null}`,
			expectedValues: []string{"a", "b", "c"},
		},
		{
			input:          `{"c": null, "b": null, "a": null}`,
			expectedValues: []string{"a", "b", "c"},
		},
		{
			input:          `{"c": 1, "b": 2, "a": 3}`,
			expectedValues: []string{"a", "b", "c"},
		},
		{
			input:     `"abc"`,
			expectErr: true,
		},
		{
			input:     `{`,
			expectErr: true,
		},
		{
			input:     `[`,
			expectErr: true,
		},
	}

	for idx := range cases {
		t.Run(fmt.Sprintf("case #%d", idx), func(t *testing.T) {
			c := cases[idx]

			var v strListOrMap
			err := json.Unmarshal([]byte(c.input), &v)
			if c.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expectedValues, v.Values())
			}
		})
	}
}
