package presenter

import (
	"github.com/Azure/ShieldGuard/sg/internal/engine"
	"github.com/Azure/ShieldGuard/sg/internal/policy"
	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/utils"
)

type policyRuleObj struct {
	Name    string `json:"name" yaml:"name"`
	DocLink string `json:"doc_link,omitempty" yaml:"doc_link,omitempty"`
}

func asPolicyRuleObj(rule policy.Rule, docLink string) policyRuleObj {
	return policyRuleObj{
		Name:    rule.Name,
		DocLink: docLink,
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
		Rule:     asPolicyRuleObj(result.Rule, result.RuleDocLink),
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
		Failures:   utils.Map(queryResult.Failures, asResultObj),
		Warnings:   utils.Map(queryResult.Warnings, asResultObj),
		Exceptions: utils.Map(queryResult.Exceptions, asResultObj),
	}
}

func asQueryResultsObjList(queryResultsList []result.QueryResults) []queryResultsObj {
	return utils.Map(queryResultsList, asQueryResultsObj)
}
