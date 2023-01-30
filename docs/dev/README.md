# Contributing Quick Start

## Repo Setup

This repo contains the source code of ShieldGuard tools and well-known policies.
We organize them in a sort-of "monorepo" approach:

- `docs/`: public facing documentations
- `policies/`: well-known, reusable policy packages
- `sg/`: `sg` CLI implementation

## `sg`: ShieldGuard CLI

`sg` is a CLI application written with Go, based on `OPA` and `Conftest`. To start hacking
on this tool, we should have a Go environment ready. We can simply building the binary with:

```
$ cd sg
$ make build
go fmt ./...
go vet ./...
go build -o bin/sg ./cmd/sg
```

This command builds the `sg` to the `bin/` folder from source code:

```
$ ./bin/sg -h
<omitted help message>
```

We have a few others hacking commands defined in the `Makefile`, which we can get the full help by:

```
$ make
Usage:
  make <target>

General
  help             Display this help.

Development
  fmt              Run go fmt against code.
  vet              Run go vet against code.
  test             Run go test against code.
  lint             Run linters against code.
  tidy             Tidy go.mod and go.sum.

Build
  build            Build the project.
  build-sg         Build the sg binary.
```

For example, we can run unit test by: `make test`:

```
$ make test
<omitted unit test output>
```

These commands should cover most of the development scenarios.

We also have Godoc published in page: https://pkg.go.dev/github.com/Azure/ShieldGuard/sg .

## Policy Guides

TODO(hbc): document how to write and reuse well-known policies

<!--
TODO(hbc): complete this part of docs

## Cookbook

### How to...

- add a new presenter?
- make a release?

-->