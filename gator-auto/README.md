# gator-auto

## Rationale

[Gator][gator] provides a way to run gatekeeper policies locally. This is a helpful addition for shift-left detection.
However, we identified a few shortcomings while integrating it in our daily usage:

- Gator requires to define the rego policy in Kubernetes object manifests. This sometime makes the policy code hard to read and review.
- Gator supports only generated Kubernetes manifests. It's very common that we manages the manifests with Kustomize / Helm in codebase. Hence, there is a gap to apply gator to the source files.
- Gator outputs doesn't correlate to source file. When using gator in local and CI/CD environments, showing the detected issues alongside with the source file information makes it easier to correlate to the source.

`gator-auto` is a PoC of addressing above issues leveraging gator's core functionality.

[gator]: https://open-policy-agent.github.io/gatekeeper/website/docs/gator/

## Building

```
$ git clone https://github.com/Azure/ShieldGuard.git
$ cd ShieldGuard
$ git checkout gator-auto
$ make build
$ /bin/gator-auto -h
Usage:
  gator-auto [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  test

Flags:
  -h, --help   help for gator-auto

Use "gator-auto [command] --help" for more information about a command.
```

## Usage

Examples can be found in the [examples](./examples/) folder:

```
./examples
├── helm            # helm example
├── kustomize       # kustomize example
├── plain           # Kubernetes manifests without templates
└── policies        # Gatekeeper policy
```

### Policy Definition

We define a sample policy for checking required annotations from object spec:

`001-required-annotations.rego`

```
# METADATA
# custom:
#  enforcementAction: deny
package requiredannotations

import future.keywords.if

violation[{"msg": msg}] {
	is_target

	required_annotations := ["acme/owning-team", "acme/owning-contact"]

	required_annotation := required_annotations[_]
	not input.review.object.metadata.annotations[required_annotation]
	msg := sprintf(
		"%s/%s missing required annotation: %s",
		[
			input.review.kind.kind,
			input.review.object.metadata.name,
			required_annotation,
		],
	)
}

# ... omitted the remaining code
```

### Plain Usage

```
$ ./bin/gator-auto test --filename examples/plain/aks-quickstart.yaml --policy examples/policies
```

<details>
<summary>
output
</summary>

```
examples/plain/aks-quickstart.yaml FAIL Deployment/order-service missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Deployment/order-service missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Deployment/product-service missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Deployment/product-service missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Deployment/rabbitmq missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Deployment/rabbitmq missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Deployment/store-front missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Deployment/store-front missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Service/order-service missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Service/order-service missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Service/product-service missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Service/product-service missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Service/rabbitmq missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Service/rabbitmq missing required annotation: acme/owning-team
examples/plain/aks-quickstart.yaml FAIL Service/store-front missing required annotation: acme/owning-contact
examples/plain/aks-quickstart.yaml FAIL Service/store-front missing required annotation: acme/owning-team
```

</details>

### Kustomize Usage

We defined a few overlays under the kustomize folder:

```
./examples/kustomize
└── overlays
    ├── base            # base overlay, contains the base spec
    ├── dev             # dev env overlay. It adds `acme/owning-team` annotation via `commonAnnotation`
    └── prod            # prod env overlay. It adds `acme/owning-contact` annotation via `commonAnnotation`
```

```
$ ./bin/gator-auto test --kustomize examples/kustomize/overlays/dev --policy examples/policies
$ ./bin/gator-auto test --kustomize examples/kustomize/overlays/prod --policy examples/policies
```

<details>
<summary>
output
</summary>

For `dev` overlay:

```
examples/kustomize/overlays/base/order-service-deployment.yaml FAIL Deployment/order-service missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/order-service-deployment.yaml FAIL Service/order-service missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/product-service-deployment.yaml FAIL Deployment/product-service missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/product-service-deployment.yaml FAIL Service/product-service missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/rabbitmq-deployment.yaml FAIL Deployment/rabbitmq missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/rabbitmq-deployment.yaml FAIL Service/rabbitmq missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/store-service-deployment.yaml FAIL Deployment/store-front missing required annotation: acme/owning-contact
examples/kustomize/overlays/base/store-service-deployment.yaml FAIL Service/store-front missing required annotation: acme/owning-contact
```

For `prod` overlay:

```
examples/kustomize/overlays/base/order-service-deployment.yaml FAIL Deployment/order-service missing required annotation: acme/owning-team
examples/kustomize/overlays/base/order-service-deployment.yaml FAIL Service/order-service missing required annotation: acme/owning-team
examples/kustomize/overlays/base/product-service-deployment.yaml FAIL Deployment/product-service missing required annotation: acme/owning-team
examples/kustomize/overlays/base/product-service-deployment.yaml FAIL Service/product-service missing required annotation: acme/owning-team
examples/kustomize/overlays/base/rabbitmq-deployment.yaml FAIL Deployment/rabbitmq missing required annotation: acme/owning-team
examples/kustomize/overlays/base/rabbitmq-deployment.yaml FAIL Service/rabbitmq missing required annotation: acme/owning-team
examples/kustomize/overlays/base/store-service-deployment.yaml FAIL Deployment/store-front missing required annotation: acme/owning-team
examples/kustomize/overlays/base/store-service-deployment.yaml FAIL Service/store-front missing required annotation: acme/owning-team
```

</details>

### Helm Usage

`gator-auto` supports rendering Helm chart using chart's default `values.yaml` settings.
The sample Helm chart defined the required annotations for the `rabbitmq` component.

```
$ ./bin/gator-auto test --helm-chart examples/helm/aks-quickstart --policy examples/policies
```

<details>
<summary>
output
</summary>

```
aks-quickstart/templates/order-service-deployment.yaml FAIL Deployment/order-service missing required annotation: acme/owning-contact
aks-quickstart/templates/order-service-deployment.yaml FAIL Deployment/order-service missing required annotation: acme/owning-team
aks-quickstart/templates/order-service-deployment.yaml FAIL Service/order-service missing required annotation: acme/owning-contact
aks-quickstart/templates/order-service-deployment.yaml FAIL Service/order-service missing required annotation: acme/owning-team
aks-quickstart/templates/product-service-deployment.yaml FAIL Deployment/product-service missing required annotation: acme/owning-contact
aks-quickstart/templates/product-service-deployment.yaml FAIL Deployment/product-service missing required annotation: acme/owning-team
aks-quickstart/templates/product-service-deployment.yaml FAIL Service/product-service missing required annotation: acme/owning-contact
aks-quickstart/templates/product-service-deployment.yaml FAIL Service/product-service missing required annotation: acme/owning-team
aks-quickstart/templates/store-service-deployment.yaml FAIL Deployment/store-front missing required annotation: acme/owning-contact
aks-quickstart/templates/store-service-deployment.yaml FAIL Deployment/store-front missing required annotation: acme/owning-team
aks-quickstart/templates/store-service-deployment.yaml FAIL Service/store-front missing required annotation: acme/owning-contact
aks-quickstart/templates/store-service-deployment.yaml FAIL Service/store-front missing required annotation: acme/owning-team
```

</details>
