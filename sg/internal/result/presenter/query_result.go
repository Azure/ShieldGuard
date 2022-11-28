package presenter

import (
	"github.com/Azure/ShieldGuard/sg/internal/engine"
	"github.com/Azure/ShieldGuard/sg/internal/policy"
	"github.com/Azure/ShieldGuard/sg/internal/result"
)

type policyRuleObj struct {
	Name string `json:"name" yaml:"name"`
}

func asPolicyRuleObj(rule policy.Rule) policyRuleObj {
	return policyRuleObj{
		Name: rule.Name,
	}
}

type resultObj struct {
	Query    string                 `json:"query" yaml:"query"`
	Rule     policyRuleObj          `json:"rule" yaml:"rule"`
	Message  string                 `json:"message" yaml:"message"`
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func asResultObj(result result.Result) resultObj {
	return resultObj{
		Query:    result.Query,
		Rule:     asPolicyRuleObj(result.Rule),
		Message:  result.Message,
		Metadata: result.Metadata,
	}
}

type queryResultsObj struct {
	Filename   string      `json:"filename" yaml:"filename"`
	Namespace  string      `json:"namespace" yaml:"namespace"`
	Success    int         `json:"success" yaml:"success"`
	Failures   []resultObj `json:"failures" yaml:"failures"`
	Warnings   []resultObj `json:"warnings" yaml:"warnings"`
	Exceptions []resultObj `json:"exceptions" yaml:"exceptions"`
}

func asQueryResultsObj(queryResult result.QueryResults) queryResultsObj {
	return queryResultsObj{
		Filename:   queryResult.Source.Name(),
		Namespace:  engine.PackageMain,
		Success:    queryResult.Successes,
		Failures:   mapList(queryResult.Failures, asResultObj),
		Warnings:   mapList(queryResult.Warnings, asResultObj),
		Exceptions: mapList(queryResult.Exceptions, asResultObj),
	}
}

func asQueryResultsObjList(queryResultsList []result.QueryResults) []queryResultsObj {
	return mapList(queryResultsList, asQueryResultsObj)
}

func mapList[T any, U any](list []T, fn func(T) U) []U {
	result := make([]U, len(list))
	for i, item := range list {
		result[i] = fn(item)
	}
	return result
}
