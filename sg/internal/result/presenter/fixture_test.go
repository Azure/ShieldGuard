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
					Message:     "fail message1",
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/001-rego.md",
					Rule: policy.Rule{
						Kind: policy.QueryKindDeny,
						Name: "001-rule",
					},
				},
				{
					Message:     "fail message2",
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/002-rego.md",
					Rule: policy.Rule{
						Kind: policy.QueryKindDeny,
						Name: "002-rule",
					},
				},
			},
			Warnings: []result.Result{
				{
					Message:     "warn message1",
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/001-rego.md",
					Rule: policy.Rule{
						Kind: policy.QueryKindWarn,
						Name: "001-rule",
					},
				},
				{
					Message:     "warn message2",
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/002-rego.md",
					Rule: policy.Rule{
						Kind: policy.QueryKindWarn,
						Name: "002-rule",
					},
				},
			},
			Exceptions: []result.Result{
				{
					Message:     "",
					RuleDocLink: "https://github.com/Azure/ShieldGuard/docs/003-rego.md",
					Rule: policy.Rule{
						Kind: policy.QueryKindException,
						Name: "003-rule",
					},
				},
			},
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
