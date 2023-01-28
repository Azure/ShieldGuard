# Get Started

## Installation

ShieldGuard provides a CLI tool `sg` for running policy checks.

### From release artifacts

`sg` is available on Linux, Mac OS and Windows. We can download the pre-built binaries
from [release page][sg_release_page].

[sg_release_page]: https://github.com/Azure/ShieldGuard/releases

<!-- TODO(hbc): add download script for downloading specific version -->

### From source code

`sg` is written in Go. We can install it from source code:

```
$ go install github.com/Azure/ShieldGuard/sg@latest
```

### Validating the binary

Once we successfully installed `sg`, we can invoke as:

```
$ sg -h
Enables best security practices for your project from day zero.

Usage:
  sg [command]

Available Commands:
  help        Help about any command
  test        Test targets under the project.

Flags:
  -h, --help   help for sg

Use "sg [command] --help" for more information about a command.
```

## Running First Policy Check

Policy checks are expressed using [Rego language][rego], which are compiled and executed by 
[Open Policy Agent (OPA)][opa].

[rego]: https://www.openpolicyagent.org/docs/latest/policy-language/
[opa]: https://github.com/open-policy-agent/opa

We can use Rego to express a set of expectations to a collection of data.
In this tutorial, we will validate if a Kubernetes deployment spec is setting a required label `app/name=<app-name-value>`.

The full sample and code can be found from [./samples/001-get-started](./samples/001-get-started).

### Bootstrap Policy Project

In ShieldGuard, policies and data are organized into a `project`: each project contains *one ore more data targets*  and *zero or more policies*. These mappings are defined with `sg-project.yaml`.

We can bootstrap a test project as follow:

```
$ mkdir -p 001-get-started/data 001-get-started/policy
$ cat <<EOF > 001-get-started/sg-project.yaml
files:
- name: 001-get-started
  paths:
  - data
  policies:
  - policy
EOF
```

> :information_source: In this project, we will store the data under the `data` folder, while putting policies under the `policy` folder. We can update to different layouts in real world projects by specifying in the `sg-project.yaml`.

### Creating Test Data

Now, we will create two example deployment specs with following content:

```
$ cat <<EOF > 001-get-started/data/mysql-app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql
  labels:
    app/name: mysql
spec:
  selector:
    matchLabels:
      component: mysql
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        component: mysql
    spec:
      containers:
      - image: mysql:5.6
        name: mysql
EOF
$ cat <<EOF > 001-get-started/data/postgres-app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
spec:
  selector:
    matchLabels:
      component: postgres
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        component: postgres
    spec:
      containers:
      - image: postgres:15.1
        name: postgres
EOF
```

> :information_source: In above two samples, the mysql deployment has the required label `app/name=mysql`, while postgres one is missing. We expect to use `sg` to detect such issue in this tutorial.

### Writing the Policy

```
$ cat <<EOF > 001-get-started/policy/001-app-name-required.rego
package main

deny_app_name_required[msg] {
	input.apiVersion
	input.metadata
	not input.metadata.labels["app/name"]
	msg := sprintf("app name label app/name is required for %s/%s .", [input.kind, input.metadata.name])
}
EOF
```

> :information_source: To learn more about Rego/OPA, please refer to the [official guide][opa_official_guide].

[opa_official_guide]: https://www.openpolicyagent.org/docs/latest/policy-reference/

### Executing Test

With all the test data and policies ready, now we can invoke the test on this project:

```
$ sg test -c 001-get-started/sg-project.yaml ./001-get-started -o json
[
  {
    "filename": "data/mysql-app-deployment.yaml",
    "namespace": "main",
    "success": 1,
    "failures": [],
    "warnings": [],
    "exceptions": []
  },
  {
    "filename": "data/postgres-app-deployment.yaml",
    "namespace": "main",
    "success": 0,
    "failures": [
      {
        "query": "data.main.deny_app_name_required",
        "rule": {
          "name": "app_name_required"
        },
        "message": "app name label app/name is required for Deployment/postgres ."
      }
    ],
    "warnings": [],
    "exceptions": []
  }
]
Error: test failed: found 1 failure(s), 0 warning(s)
```

As we can see, `sg` failed because of the postgres deployment is missing the required `app/name` label.

We can fix it by adding the label:

```
cat <<EOF > 001-get-started/data/postgres-app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  labels:
    app/name: postgres
spec:
  selector:
    matchLabels:
      component: postgres
  strategy:
    type: Recreate
  template:
    metadata:
      labels:
        component: postgres
    spec:
      containers:
      - image: postgres:15.1
        name: postgres
EOF
```

Then re-run the test command:

```
$ sg test -c 001-get-started/sg-project.yaml ./001-get-started -o json
[
  {
    "filename": "data/mysql-app-deployment.yaml",
    "namespace": "main",
    "success": 1,
    "failures": [],
    "warnings": [],
    "exceptions": []
  },
  {
    "filename": "data/postgres-app-deployment.yaml",
    "namespace": "main",
    "success": 1,
    "failures": [],
    "warnings": [],
    "exceptions": []
  }
]
```

That's it! You have successfully defined and ran the test in your project. For next steps, please checkout our [other guides](./README.md).