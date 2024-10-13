package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/openai/openai-go"
	"github.com/stretchr/testify/assert"
)

func Test_ToAgentFunctionOrPanic(t *testing.T) {
	f := ToAgentFunctionOrPanic(
		"test",
		"test function. First argument is foo, second argument is bar.",
		func(ctx context.Context, agentCtx AgentContext, foo string, bar string) (bool, error) {
			t.Log("foo:", foo)
			t.Log("bar:", bar)

			return false, nil
		},
	)

	toolParam := f.ToolParam()
	assert.Equal(t, openai.ChatCompletionToolTypeFunction, toolParam.Type.Value)
	assert.Equal(t, "test", toolParam.Function.Value.Name.Value)
	assert.Equal(t, "test function. First argument is foo, second argument is bar.", toolParam.Function.Value.Description.Value)
	// TODO: check arguments

	fmt.Println(f.ToolParam())
	rv, err := f.Invoke(context.Background(), AgentContext{}, json.RawMessage(`{"arg0": "foo_value", "arg1": "bar_value"}`))
	assert.NoError(t, err)
	assert.NotNil(t, rv)
	assert.Equal(t, "false", rv.Value)
}
