# Writing Policy

ShieldGuard validates various kind of data by executing a collection of defined policy rules.
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
loaded_input_data = load_all_input_data()

for policy_rule in gather_all_rules():
  for input_data in loaded_input_data:
    if rego.query(input_data, policy_rule.rule_name) == MATCH:
      if rego.query(input_data, exception[_][_] == policy_rule.rule_name) != MATCH:
        switch policy_rule.kind:
          case "DENY" # rule == MATCH, rule is defined with DENY kind
            deny_results.add(policy_rule)
          case "WARN" # rule == MATCH, rule is defined with WARN kind
            warn_results.add(policy_rule)
      else: # exception == MATCH, there is at least 1 exception excludes this violating rule
        exception_results.add(policy_rule)
    else: # query result != MATCH. All validations pass
      success_rules_count += 1
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
input.kind # resolves to "Namespace"
input.meatadata.name # resolves to "my-app"
input.metadata.labels.tier # resolves to "production"
```

A typical rule implementation flow would be:

1. check for "signature" from the input data structure. For instance, fields with targeted values;
2. validate against interested fields and values;
3. gather results based on the check. If a `true` value is calculated, then the rule engine collects the rule with associated kind.

Using the above namespace data, if we want to enforce setting the "owner" label for it, we can use following code:

```rego
warn_missing_owner_label[msg] {
  input.kind == "Namespace" # check for signature
  input.metadata.labels.tier == "production" # check for signature
  not input.metadata.labels.owner # validate against interested fields and values
  msg := "missing owner label for production namespace" # generate result
}
```

[rego_policy_lang]: https://www.openpolicyagent.org/docs/latest/policy-language/

### Policy Documentation

Even though each rule can provide an advisory message to help configuration authors to understand why one or more rules have failed during the execution, however, sometimes it's still challenge to provide detailed background, explanations and mitigation steps. Therefore, in ShieldGuard, we prompt the documentation with higher priority: each rule comes with an optional documentation. These documentations can be referenced via a URL, which is defined in the policy package settings.

<details>

<summary>Documentation Example</summary>

Suppose we have a policy rule implementation using `001-missing_owner_label.rego`

```rego
# 001-missing_owner_label.rego

warn_missing_owner_label[msg] { /* implementation details */ }
```

We can create a companion markdown doc: `001-missing_owner_label.md`

```md
# Rule: 001 Miss Owner Label (`missing_owner_label`)

## Description

All production tier apps should set owner label.

## Mitigation

Declare the owner label in the namespace object:

metadata:
  labels:
    owner: infra-team
```

</details>

## Policy Package

After writing bunch of individual policy rules, we can group them into a bigger group for reusing. In this case, we can create a policy package for these rules. A package contains two part:

1. a collection of policy rules, documentations
2. package settings

A typical package folder would look like:

```
my-package/
 /docs               # <- policy document folder
   /001-rule_foo.md
   /002-rule_bar.md
   /003-rule_baz.md
   ...
 /sg-project.yaml    # <- policy package settings
 /001-rule_foo.rego  # <- policy implementations
 /002-rule_bar.rego
 /003-rule_baz.rego
 ...
```

### Policy Name & Source File Name

It's very common that there are multiple policy rules under a policy package. To help better organizing the policy rule implementations and documentations, we suggest package authors to 1. create one rule per rego file; 2. order rules with sequence id prefix. For example:

| rule name | rule file name | doc file name |
|:----------|:---------------|:--------------|
| `rule_foo` | `001-rule_foo.rego` | `001-rule_foo.md` |
| `rule_bar` | `002-rule_foo.rego` | `002-rule_foo.md` |

With such structure, the user can easily navigate inside the policy implementations and documentations, while policy author can keep track of the policy rules easily.

### Package Settings

In a ShieldGuard package, we use `sg-project.yaml` to define the package related metadata and settings. This file should be located in the package root level. Current supported fields are:

```yaml
# sg-project.yaml

# settings for policy rule
rule:
  # doc_link specifies teh policy rule document link format.
  # The value can be formatted using Golang's text.Template. Following template variables are available:
  #
  #   - {{.Name}}: the name of the rule. Ex: `missing_owner_label`
  #   - {{.Kind}}: the kind of the rule. Ex: `deny` / `warn` 
  #   - {{.SourceFileName}}: the source file name where the rule being defined, without extension. Ex: `001-missing_owner_label` . If the rule is not defined from a source file, an empty value will be used.
  doc_link: 'https://example.com/my-policy/{{.SourceFileName}}.md'
```

### Reusing Policy Packages

TBD. We are working on remote packages support, stay tuned.