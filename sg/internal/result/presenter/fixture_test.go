package presenter

import (
	"github.com/Azure/ShieldGuard/sg/internal/policy"
	"github.com/Azure/ShieldGuard/sg/internal/result"
	"github.com/Azure/ShieldGuard/sg/internal/source/testsource"
)

func testQueryResults() []result.QueryResults {
	return []result.QueryResults{
		{
			Source: &testsource.TestSource{NameFunc: func() string {
				return "file name"
			}},
			Successes: 2,
			Failures: []result.Result{
				{
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/001-rego.md",
					Message:     "fail message1",
					Rule: policy.Rule{
						Kind: policy.QueryKindDeny,
						Name: "001-rule",
					},
				},
				{Message: "fail message2"},
			},
			Warnings: []result.Result{
				{Message: "warn message1"},
				{
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/002-rego.md",
					Message:     "warn message2",
				},
			},
			Exceptions: []result.Result{{Message: "exception message1"}},
		},
		{
			Source: &testsource.TestSource{NameFunc: func() string {
				return ""
			}},
			Successes:  0,
			Failures:   []result.Result{},
			Warnings:   []result.Result{},
			Exceptions: []result.Result{},
		},
	}
}
