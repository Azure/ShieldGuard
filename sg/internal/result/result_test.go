package result

import (
	"fmt"
	"testing"

	"github.com/Azure/ShieldGuard/sg/internal/source/testsource"
	"github.com/open-policy-agent/opa/rego"
	"github.com/stretchr/testify/assert"
)

func Test_fromMetadata(t *testing.T) {
	cases := []struct {
		query     string
		metadata  map[string]interface{}
		expected  Result
		expectErr bool
	}{
		{
			query:     "data.deny_foo",
			metadata:  map[string]interface{}{},
			expectErr: true,
		},
		{
			query: "data.deny_foo",
			metadata: map[string]interface{}{
				"msg": 123,
			},
			expectErr: true,
		},
		{
			query: "data.deny_foo",
			metadata: map[string]interface{}{
				"msg": "foo",
			},
			expected: Result{
				Query:    "data.deny_foo",
				Message:  "foo",
				Metadata: map[string]interface{}{},
			},
		},
		{
			query: "data.deny_foo",
			metadata: map[string]interface{}{
				"msg": "foo",
				"foo": "bar",
			},
			expected: Result{
				Query:   "data.deny_foo",
				Message: "foo",
				Metadata: map[string]interface{}{
					"foo": "bar",
				},
			},
		},
	}

	for idx := range cases {
		c := cases[idx]
		t.Run(fmt.Sprintf("case #%d", idx), func(t *testing.T) {
			actual, err := fromMetadata(c.query, c.metadata)
			if c.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, actual)
			}
		})
	}
}

func Test_FromRegoExpression(t *testing.T) {
	cases := []struct {
		query      string
		expression *rego.ExpressionValue
		expected   []Result
		expectErr  bool
	}{
		{
			query: "data.foo",
			expression: &rego.ExpressionValue{
				Value: "something",
			},
			expected: []Result{empty("data.foo")},
		},
		{
			query: "data.foo",
			expression: &rego.ExpressionValue{
				Value: []interface{}{"value"},
			},
			expected: []Result{fromString("data.foo", "value")},
		},
		{
			query: "data.foo",
			expression: &rego.ExpressionValue{
				Value: []interface{}{
					map[string]interface{}{},
				},
			},
			expectErr: true,
		},
		{
			query: "data.foo",
			expression: &rego.ExpressionValue{
				Value: []interface{}{
					map[string]interface{}{
						"msg": "value",
						"foo": "bar",
					},
				},
			},
			expected: []Result{
				{
					Query:   "data.foo",
					Message: "value",
					Metadata: map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	}

	for idx := range cases {
		c := cases[idx]
		t.Run(fmt.Sprintf("case #%d", idx), func(t *testing.T) {
			actual, err := FromRegoExpression(c.query, c.expression)
			if c.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, c.expected, actual)
			}
		})
	}
}

func Test_Result_Passed(t *testing.T) {
	passed := Result{}
	assert.True(t, passed.Passed())

	failed := Result{Message: "failed"}
	assert.False(t, failed.Passed())
}

func Test_QueryResults_Merge(t *testing.T) {
	left := QueryResults{
		Source: &testsource.TestSource{
			NameFunc: func() string {
				return "left"
			},
		},
		Successes: 10,
		Failures: []Result{
			{Message: "failed-left-1"},
		},
		Warnings: []Result{
			{Message: "warn-left-1"},
		},
		Exceptions: []Result{
			{Message: "exc-left-1"},
		},
	}
	right := QueryResults{
		Source: &testsource.TestSource{
			NameFunc: func() string {
				return "right"
			},
		},
		Successes: 3,
		Failures: []Result{
			{Message: "failed-right-1"},
			{Message: "failed-right-2"},
		},
		Warnings: []Result{
			{Message: "warn-right-1"},
		},
		Exceptions: []Result{
			{Message: "exc-right-1"},
		},
	}

	merged := left.Merge(right)
	assert.Equal(t, "left", merged.Source.Name())
	assert.Equal(t, 13, merged.Successes)
	assert.Len(t, merged.Failures, 3)
	assert.Len(t, merged.Warnings, 2)
	assert.Len(t, merged.Exceptions, 2)
}
