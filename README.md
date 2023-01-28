<div align="center">
<!-- TODO: logo -->
<h1>ShieldGuard</h1>
<p>Enables best security practices for your project from day zero.</p>

[![Unit Test](https://github.com/Azure/ShieldGuard/actions/workflows/unit-test.yaml/badge.svg)](https://github.com/Azure/ShieldGuard/actions/workflows/unit-test.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Azure/ShieldGuard/sg.svg)](https://pkg.go.dev/github.com/Azure/ShieldGuard/sg)
</div>

## What's ShieldGuard?

ShieldGuard is a **modular tool** and a **process** for enforcing various kind of validations on structured data. These data can be
the JSON/YAML/TOML/XML/... configurations from project source code, or the runtime data from your production environments.

## Quick Start

Get ShieldGuard running in 5 minutes: [./docs/manual/get-started.md](/docs/manual/get-started.md)

## Documentations

Interested to more usage scenarios? Checkout [./docs/manual](/docs/manual/) for more examples!

### Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

For step by step development setup & contribution guides, please see [./docs/dev](/docs/dev/) folder.

### Security

Please follow [SECURITY.md](/SECURITY.md) to report security issues.

## History and Why

ShieldGuard is a based on [Open Policy Agent (OPA)][opa] and heavily inspired by:

- [Conftest][conftest]
- [defsec][]
- ... and many other tools!

We decided to build a new tool based on following reasons:

1. ShieldGuard aims to provide a unified way to write checks using **vanilla** Rego language. This means you can reuse the checks
   without the need to depend on ShieldGuard itself;
2. ShieldGuard makes policy and check documentation as first-class citizen: it provides a convention approach for writing, organizing
   and referencing the documentations alongside with the policies;
3. ShieldGuard exposes itself via a modular types and packages, which enables further composition and building new tools easily.

[opa]: https://github.com/open-policy-agent/opa
[conftest]: https://www.conftest.dev/
[defsec]: https://github.com/aquasecurity/defsec

## License

[MIT](/LICENSE)

### Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft 
trademarks or logos is subject to and must follow 
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
