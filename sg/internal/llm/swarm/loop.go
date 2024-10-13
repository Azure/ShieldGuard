package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"reflect"
	"slices"

	"github.com/openai/openai-go"
)

type Loop struct {
	logger *slog.Logger
	client *openai.Client
}

type LoopRunParams struct {
	Messages      []openai.ChatCompletionMessageParamUnion
	Agent         Agent
	AgentContext  AgentContext
	ModelOverride string
	MaxTurns      int
	ExecuteTools  bool
}

func (l *Loop) getChatCompletion(
	ctx context.Context,
	agent Agent,
	agentContext AgentContext,
	history []openai.ChatCompletionMessageParamUnion,
	modelOverride string,
) (*openai.ChatCompletion, error) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(agent.Instructions(agentContext)),
	}
	messages = append(messages, history...)

	var tools []openai.ChatCompletionToolParam
	for _, f := range agent.Functions() {
		tools = append(tools, f.ToolParam())
	}

	params := openai.ChatCompletionNewParams{
		Messages:   openai.F(messages),
		Tools:      openai.F(tools),
		ToolChoice: openai.F(agent.ToolChoice()),
		Model:      openai.F(agent.Model()),
		// TODO: stream
	}
	if modelOverride != "" {
		params.Model = openai.F(modelOverride)
	}
	if len(tools) > 0 {
		params.ParallelToolCalls = openai.Bool(agent.ParallelToolsCall())
	}

	return l.client.Chat.Completions.New(ctx, params)
}

func (l *Loop) handleToolCalls(
	ctx context.Context,
	agentContext AgentContext,
	toolCalls []openai.ChatCompletionMessageToolCall,
	functions []AgentFunction,
) (*Response, error) {
	functionsMap := map[string]AgentFunction{}
	for _, f := range functions {
		param := f.ToolParam()
		functionsMap[param.Function.Value.Name.Value] = f
	}

	rv := &Response{}

	// TODO: record tool name in message
	for _, toolCall := range toolCalls {
		name := toolCall.Function.Name
		f, exists := functionsMap[name]
		if !exists {
			l.logger.Warn("tool called with unknown function", slog.String("function.name", name))
			message := openai.ToolMessage(toolCall.ID, fmt.Sprintf("Error: tool %q not found.", name))
			rv.Messages = append(rv.Messages, message)
			continue
		}

		arguments := json.RawMessage(toolCall.Function.Arguments)
		l.logger.Debug("invoking tool", slog.String("function.name", name), slog.String("function.arguments", string(arguments)))
		invokeResult, err := f.Invoke(ctx, agentContext, arguments)
		if err != nil {
			l.logger.Warn("error invoking tool", slog.String("function.name", name), slog.String("error", err.Error()))
			message := openai.ToolMessage(toolCall.ID, fmt.Sprintf("Error: %v", err))
			rv.Messages = append(rv.Messages, message)
			continue
		}

		rv.AgentContext = rv.AgentContext.Merge(invokeResult.AgentContext)
		toolMessageContent := invokeResult.Value
		if !isNil(invokeResult.Agent) {
			agentName := invokeResult.Agent.Name()
			value, err := json.Marshal(map[string]any{"assistant": agentName})
			if err != nil {
				return nil, fmt.Errorf("error marshalling agent name: %q", agentName)
			}
			toolMessageContent = string(value)
			rv.Agent = invokeResult.Agent
		}

		message := openai.ToolMessage(toolCall.ID, toolMessageContent)
		rv.Messages = append(rv.Messages, message)
	}

	return rv, nil
}

func (l *Loop) Run(
	ctx context.Context,
	params LoopRunParams,
) (*Response, error) {
	maxTurns := params.MaxTurns
	if maxTurns == 0 {
		maxTurns = math.MaxInt64
	}

	activeAgent := params.Agent
	agentContext := params.AgentContext.Clone()
	history := slices.Clone(params.Messages)
	initLen := len(history)

	for {
		if len(history) >= maxTurns {
			break
		}
		if isNil(activeAgent) {
			break
		}

		completion, err := l.getChatCompletion(
			ctx, activeAgent, agentContext, history, params.ModelOverride,
		)
		if err != nil {
			return nil, err
		}
		if len(completion.Choices) == 0 {
			return nil, fmt.Errorf("no choices in completion")
		}

		message := completion.Choices[0].Message
		l.logger.Debug("received completion", slog.String("message.raw", message.JSON.RawJSON()))
		// TODO: record agent name with message
		history = append(history, message)

		if len(message.ToolCalls) < 1 && !params.ExecuteTools {
			l.logger.Debug("no tool calls in completion, stopping")
			break
		}

		partialResponse, err := l.handleToolCalls(ctx, agentContext, message.ToolCalls, activeAgent.Functions())
		if err != nil {
			return nil, err
		}

		history = append(history, partialResponse.Messages...)
		agentContext = agentContext.Merge(partialResponse.AgentContext)
		if partialResponse.Agent != nil {
			// switch to new agent
			activeAgent = partialResponse.Agent
		}

		// new turn...
	}

	rv := &Response{
		Messages:     history[initLen:],
		Agent:        activeAgent,
		AgentContext: agentContext,
	}
	return rv, nil
}

// TODO: stream

func isNil(v any) bool {
	if v == nil {
		return true
	}

	switch reflect.ValueOf(v).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Slice:
		return reflect.ValueOf(v).IsNil()
	default:
		return false
	}
}
