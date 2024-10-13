package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/openai/openai-go"
	"github.com/tidwall/gjson"
)

type agentFunction struct {
	toolParam openai.ChatCompletionToolParam
	invoke    func(
		ctx context.Context,
		agentCtx AgentContext,
		arguments json.RawMessage,
	) (*AgentResult, error)
}

var _ AgentFunction = (*agentFunction)(nil)

func agentFunctionFromValue(name string, description string, value any) *agentFunction {
	return &agentFunction{
		toolParam: openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String(name),
				Description: openai.String(description),
			}),
		},
		invoke: func(_ context.Context, _ AgentContext, _ json.RawMessage) (*AgentResult, error) {
			rv := &AgentResult{Value: fmt.Sprint(value)}
			return rv, nil
		},
	}
}

func (f *agentFunction) ToolParam() openai.ChatCompletionToolParam {
	return f.toolParam
}

func (f *agentFunction) Invoke(
	ctx context.Context,
	agentCtx AgentContext,
	arguments json.RawMessage,
) (*AgentResult, error) {
	return f.invoke(ctx, agentCtx, arguments)
}

// FIXME: find a library for this...
func goTypeToJSONSchemaType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "object"
	}
}

func goFunctionToAgentFunction(
	name string,
	description string,
	fValue reflect.Value,
	fType reflect.Type,
) (*agentFunction, error) {
	// supported signatures:
	//
	// func(ctx context.Context, agentCtx AgentContext, args ...any) (RV, error)
	// func(ctx context.Context, args ...any) (RV, error)
	// func(ctx context.Context) (RV, error)
	// func(agentCtx AgentContext, args ...any) (RV, error)
	// func(agentCtx AgentContext) (RV, error)
	// func(args ...any) (RV, error)

	// validating arguments...
	var (
		hasCtxArg      bool
		hasAgentCtxArg bool
	)
	argSchemas := map[string]map[string]any{}
	for i := 0; i < fType.NumIn(); i++ {
		arg := fType.In(i)
		if arg == reflect.TypeOf((*context.Context)(nil)).Elem() {
			switch i {
			case 0:
				hasCtxArg = true
				continue
			default:
				return nil, fmt.Errorf("context.Context argument must be the first argument")
			}
		}
		if arg == reflect.TypeOf((*AgentContext)(nil)).Elem() {
			switch i {
			case 0:
				hasCtxArg = false
				hasAgentCtxArg = true
				continue
			case 1:
				if !hasCtxArg {
					// TODO: enhance error message...
					return nil, fmt.Errorf("context.Context argument must be the first argument")
				}
				hasAgentCtxArg = true
				continue
			default:
				return nil, fmt.Errorf("AgentContext argument must be the first or second argument")
			}
		}

		argName := fmt.Sprintf("arg%d", len(argSchemas))
		argSchemas[argName] = map[string]any{
			"type": goTypeToJSONSchemaType(arg),
			// TODO: add description...
			"description": fmt.Sprintf("argument %d", i),
		}
	}

	// validating return values...

	if fType.NumOut() != 2 {
		return nil, fmt.Errorf("function must return exactly 2 values, got %d", fType.NumOut())
	}
	if fType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		return nil, fmt.Errorf("function must return error as the last value")
	}

	toolParam := openai.ChatCompletionToolParam{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String(name),
			Description: openai.String(description),
			Parameters: openai.F(openai.FunctionParameters{
				"type":       "object",
				"properties": argSchemas,
			}),
		}),
	}

	invoke := func(
		ctx context.Context,
		agentCtx AgentContext,
		arguments json.RawMessage,
	) (*AgentResult, error) {
		var invokeArgs []reflect.Value
		if hasCtxArg {
			invokeArgs = append(invokeArgs, reflect.ValueOf(ctx))
		}
		if hasAgentCtxArg {
			invokeArgs = append(invokeArgs, reflect.ValueOf(agentCtx))
		}
		for argName, argSchema := range argSchemas {
			v := gjson.GetBytes(arguments, argName)
			var argValue reflect.Value
			switch argSchema["type"] {
			case "string":
				argValue = reflect.ValueOf(v.String())
			case "boolean":
				argValue = reflect.ValueOf(v.Bool())
			case "integer":
				argValue = reflect.ValueOf(v.Int())
			case "number":
				argValue = reflect.ValueOf(v.Float())
			case "array":
				var vv []any
				if err := json.Unmarshal([]byte(v.Raw), &vv); err != nil {
					return nil, fmt.Errorf("error unmarshalling array argument %q: %v", argName, err)
				}
				argValue = reflect.ValueOf(vv)
			case "object":
				var vv map[string]any
				if err := json.Unmarshal([]byte(v.Raw), &vv); err != nil {
					return nil, fmt.Errorf("error unmarshalling object argument %q: %v", argName, err)
				}
				argValue = reflect.ValueOf(vv)
			}

			invokeArgs = append(invokeArgs, argValue)
		}

		invokeRets := fValue.Call(invokeArgs)
		if len(invokeRets) != 2 {
			return nil, fmt.Errorf("function must return exactly 2 values, got %d", len(invokeRets))
		}
		if !invokeRets[1].IsNil() {
			return nil, invokeRets[1].Interface().(error)
		}

		// special case for AgentResult...
		if invokeRets[0].Type() == reflect.TypeOf((*AgentResult)(nil)) {
			return invokeRets[0].Interface().(*AgentResult), nil
		}

		rv := &AgentResult{
			Value: fmt.Sprint(invokeRets[0].Interface()),
		}
		return rv, nil
	}

	return &agentFunction{
		toolParam: toolParam,
		invoke:    invoke,
	}, nil
}

func ToAgentFunctionOrPanic(name string, description string, f any) AgentFunction {
	vf := reflect.ValueOf(f)
	switch vf.Kind() {
	case reflect.Func:
		rv, err := goFunctionToAgentFunction(name, description, vf, vf.Type())
		if err != nil {
			panic(err.Error())
		}
		return rv
	case reflect.String: // TODO: create immediate function for every primitives type
		return agentFunctionFromValue(name, description, f)
	default:
		// TODO: enhance error message...
		panic(fmt.Sprintf("unsupported value type provided: %T", f))
	}
}

func AgentToFunction(a Agent) AgentFunction {
	agentName := a.Name()

	return &agentFunction{
		toolParam: openai.ChatCompletionToolParam{
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.String(a.Name()),
				Description: openai.String(a.Description()),
			}),
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
		},
		invoke: func(ctx context.Context, agentCtx AgentContext, arguments json.RawMessage) (*AgentResult, error) {
			value, err := json.Marshal(map[string]any{"assistant": agentName})
			if err != nil {
				return nil, fmt.Errorf("error marshalling agent name: %q", agentName)
			}

			rv := &AgentResult{Value: string(value), Agent: a}
			return rv, nil
		},
	}
}
