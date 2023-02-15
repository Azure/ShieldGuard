# Writing Policy

ShieldGuard validates various kind of data by executing a set of defined policy rules.
This documentation describes how to write and organize these policy rules for a project.

A simple diagram for getting quick understanding of these entities:

![](../assets/policy-package.svg)

## What's a policy rule?

Policy rule is the minimum execution unit in ShieldGuard. The query engine queries each rule with each of the input data to gather validation results.
If the input data violates the policy, the policy should return an advisory message to help user
to understand the reason and potential mitigation steps.

In ShieldGuard, we express each policy rule as an expression with OPA policy language.

We can further break down a policy rule as follow:

```rego
deny_host_volume[msg] { /* some content */ }
^^^^ ^^^^^^^^^^^ ^^^^^^^^^^^^^^^^^^^^^^^^^
   |           |                         |---- policy implementation
   |           |---- policy name
   |
   |---- policy kind
```

And besides the OPA policy language itself, we also provide a by convention documentation reference based on the policy name.

### Policy Kind & Policy Name

We define following `kind`s as a way to tell the violation severity:

| kind | description |
|:----:|-------------|
| `deny` | This violation **must** be avoided. Such violation fails the test run by default. |
| `warn` | This violation **should** be avoided, but it doesn't fail the test run. |

Example:

```rego
deny_disallowed_caps[msg] { }
^^^^ ^^^^^^^^^^^^^^^ ----  policy name
   |---- kind
```

```
FAIL - /path/to/some/config.yaml - (disallowed_caps) Container 'foo' of Deployment 'bar' should not set `securityContext.capabilities.add`.
^^^^                                ^^^^^^^^^^^^^^^ ---- policy name
   |---- a deny violation generates a FAIL result
```

> ℹ️ By convention, the kind and rule name are separated by one and only one `_`.

### Policy Implementation

Policy authors express the policy check via the [Rego's policy language][rego_policy_lang]. In each run, ShieldGuard's rule engine iterates and parses all defined policy rules. Then it goes through all input data and executes these parsed rules one by one. Each rule will be passed with *one* input data on the execution. A simplified execute flow is:

```python
for policy_rule in gather_all_rules():
  for input_data in loaded_input_data:
    if rego.query(input_data, policy_rule.rule_name) == MATCH:
      if rego.query(input_data, exception[_][_] == policy_rule.rule_name) != MATCH:
        switch policy_rule.kind:
          case "DENY"
            deny_results.add(policy_rule)
          case "WARN"
            warn_results.add(policy_rule)
      else: # exception == MATCH
        exception_results.add(policy_rule)
    else: # query result != MATCH
      success_rules += 1
```

Inside the rule implementation body, we can reference the input data via `input` variable:

Suppose we have following input data in JSON form:

```json
{
  "kind": "Namespace",
  "metadata": {
    "name": "my-app",
    "labels": {
       "tier": "production"
    }
  }
}
```

In rego:

```rego
input.kind # "Namespace"
input.meatadata.name # "my-app"
```

A typical rule implementation flow would be:

1. check for "signature" from the input data structure. For instance, fields with targeted values;
2. validate against interested fields and values;
3. gather results based on the check. If a `true` value is calculated, then the rule engine collects the rule with associated kind.

Using the above namespace data, if we want to enforce setting the "owner" label for it, we can use following code:

```rego
warn_missing_label[msg] {
  input.kind == "Namespace" # check for signature
  input.metadata.labels.tier == "production" # check for signature
  not input.metadata.labels.owner # validate against interested fields and values
  msg := "missing owner label for production namespace" # generate result
}
```

[rego_policy_lang]: https://www.openpolicyagent.org/docs/latest/policy-language/

### Policy Documentation

## Policy Package

### Policy Name & Source File Name

It's very common that there are multiple policy rules under a policy package. To help better organizing the policy rule implementations and documentations, we suggest package authors to 1. create one rule per rego file; 2. order rules with sequence id prefix. For example:

```
my-package/
 /docs               # <- policy document folder
   /001-rule-foo.md
   /002-rule-bar.md
   /003-rule-baz.md
   ...
 /sg-project.yaml    # <- policy package settings
 /001-rule-foo.rego  # <- policy implementations
 /002-rule-bar.rego
 /003-rule-baz.rego
 ...
```

Inside each rego rule, we can implementation as follow:

```rego
# 001-rule-foo.rego

deny_rule_foo[msg] {
  # ... implementation details ...
}
```

With such structure, the user can easily navigate inside the policy implementations and documentations, while policy author can keep track of the policy rules easily.

### Reusing Policy Packages

TBD